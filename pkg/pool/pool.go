package pool

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kong/pg-aurora-client/pkg/metrics"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"time"
)

type ConnectionValidator interface {
	ValidateQuery(ctx context.Context) error
}

// ConnPool db conns pool interface
type ConnPool interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	Close() error

	Conn(ctx context.Context) (*sql.Conn, error)
	Driver() driver.Driver

	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Ping() error
	PingContext(ctx context.Context) error

	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)

	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

type PGXConnPool interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	Config() *pgxpool.Config
	Stat() *pgxpool.Stat
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) error
	BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Ping(ctx context.Context) error
}

type ValidationFunction func(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool

type AuroraPGPool struct {
	innerPoolMutex         sync.RWMutex
	innerPool              *pgxpool.Pool
	queryValidationFunc    ValidationFunction
	logger                 *zap.Logger
	queryHealthCheckPeriod time.Duration
	closeChan              chan struct{}
	closeOnce              sync.Once
}

func (p *AuroraPGPool) Close() {
	p.closeOnce.Do(func() {
		close(p.closeChan)
		p.innerPoolMutex.Lock()
		p.innerPool.Close()
		p.innerPoolMutex.Unlock()
	})
}

func (p *AuroraPGPool) backgroundQueryHealthCheck() {
	ticker := time.NewTicker(p.queryHealthCheckPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-p.closeChan:
			p.logger.Info("backgroundQueryHealthCheck exited..")
			return
		case <-ticker.C:
			p.checkQueryHealth()
		}
	}
}

func (p *AuroraPGPool) checkQueryHealth() {
	ctx := context.Background()
	stats := p.Stat()
	host := p.Config().ConnConfig.Host
	p.logger.Info("pool stats", zap.Int64("acquired", int64(stats.AcquiredConns())),
		zap.Int64("idle", int64(stats.IdleConns())),
		zap.Int64("max", int64(stats.MaxConns())))

	sendPoolConnMetrics(stats, host)
	conns := p.AcquireAllIdle(ctx)
	availableCount := len(conns)
	if availableCount == 0 { //TODO: Retry logic
		p.logger.Warn("Health check reported no available connections")
		return
	}

	p.logger.Info("started checkQueryHealth run..")
	destroyCount := 0
	if p.queryValidationFunc != nil {
		for _, conn := range conns {
			validated := p.queryValidationFunc(ctx, conn, p.logger)
			if !validated {

				err := conn.Conn().Close(ctx)
				if err != nil {
					p.logger.Warn("Invalid Connection close operation resulted in error", zap.Error(err))
				}
				destroyCount++
				p.logger.Sugar().Errorf("Connection validation healthcheck failed. destroyCount=%d",
					destroyCount)
			}
			conn.Release()
		}
	}
	p.logger.Sugar().Infof("destroyCount=%d", destroyCount)
	p.logger.Info("Connections pool state", zap.String("pg_host", host),
		zap.Int("availableCount:", availableCount), zap.Int("destroyed", destroyCount))

	// Min 2 connections out of 3
	if availableCount > 3 && destroyCount > 2 {
		p.logger.Warn("Destroying pool since ratio of destroyed connections > 0.5")
		// this means more than 50% un-leased connections have a problem and some are leased out
		p.logger.Warn("There may be straggling bad connections in the pool, trying to destroy the pool")
		pool, err := pgxpool.ConnectConfig(ctx, p.innerPool.Config())
		recreateFail := false
		if err != nil {
			p.logger.Error("Failed to recreate innerPool", zap.Error(err))
			recreateFail = true
		}
		if !recreateFail {
			err = pool.Ping(ctx)
			if err != nil {
				p.logger.Error("Failed to ping innerPool", zap.Error(err))
				recreateFail = true
			}
		}
		if !recreateFail {
			p.innerPoolMutex.Lock()
			defer p.innerPoolMutex.Unlock()
			tempPool := p.innerPool
			p.innerPool = pool // set it to new before closing the retired pool
			p.logger.Info("Pool recreated")
			tempPool.Close() // close the old connections gracefully
			go metrics.Count("pg_aurora_custom_db_destroy_count",
				1, metrics.Tag{Key: "pg_host", Value: host})
		}
	}
	p.logger.Info("ended checkQueryHealth run..")
}

func sendPoolConnMetrics(stats *pgxpool.Stat, host string) {
	metrics.Count("pg_aurora_custom_idle_conn", int64(stats.IdleConns()),
		metrics.Tag{Key: "pg_host", Value: host})
	metrics.Count("pg_aurora_custom_acquired_conn", int64(stats.AcquiredConns()),
		metrics.Tag{Key: "pg_host", Value: host})
	metrics.Count("pg_aurora_custom_max_conn", int64(stats.MaxConns()),
		metrics.Tag{Key: "pg_host", Value: host})
}

func (p *AuroraPGPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.Acquire(ctx)
}

func (p *AuroraPGPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.AcquireFunc(ctx, f)
}

func (p *AuroraPGPool) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.AcquireAllIdle(ctx)
}

func (p *AuroraPGPool) Config() *pgxpool.Config {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.Config()
}

func (p *AuroraPGPool) Stat() *pgxpool.Stat {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.Stat()
}

func (p *AuroraPGPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.Exec(ctx, sql, arguments...)
}

func (p *AuroraPGPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.Query(ctx, sql, args...)
}

func (p *AuroraPGPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.QueryRow(ctx, sql, args...)
}

func (p *AuroraPGPool) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{},
	f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.QueryFunc(ctx, sql, args, scans, f)
}

func (p *AuroraPGPool) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.SendBatch(ctx, b)
}

func (p *AuroraPGPool) Begin(ctx context.Context) (pgx.Tx, error) {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.Begin(ctx)
}

func (p *AuroraPGPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.BeginTx(ctx, txOptions)
}

func (p *AuroraPGPool) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.BeginFunc(ctx, f)
}

func (p *AuroraPGPool) BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.BeginTxFunc(ctx, txOptions, f)
}

func (p *AuroraPGPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string,
	rowSrc pgx.CopyFromSource) (int64, error) {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (p *AuroraPGPool) Ping(ctx context.Context) error {
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	return p.innerPool.Ping(ctx)
}

func (p *AuroraPGPool) ValidateQuery(ctx context.Context) error {
	if p.queryValidationFunc == nil {
		return errors.New("no QueryValidationFunc set")
	}
	p.innerPoolMutex.RLock()
	defer p.innerPoolMutex.RUnlock()
	conn, err := p.innerPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return err
	}
	validated := p.queryValidationFunc(ctx, conn, p.logger)
	if !validated {
		return errors.New("query validation failed")
	}
	return nil
}

func NewAuroraPool(ctx context.Context, config *Config, logger *zap.Logger) (*AuroraPGPool, error) {
	dbpool, err := pgxpool.ConnectConfig(ctx, config.PGXConfig)
	if err != nil {
		return nil, err
	}
	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	if reflect.ValueOf(config.QueryHealthCheckPeriod).IsZero() {
		config.QueryHealthCheckPeriod = defaultQueryHealthCheckPeriod
	}

	p := &AuroraPGPool{
		innerPool:              dbpool,
		logger:                 logger,
		queryValidationFunc:    config.QueryValidator,
		queryHealthCheckPeriod: config.QueryHealthCheckPeriod,
		closeChan:              make(chan struct{}),
	}
	// Start the validator
	if config.QueryValidator != nil {
		go p.backgroundQueryHealthCheck()
	}
	return p, nil
}
