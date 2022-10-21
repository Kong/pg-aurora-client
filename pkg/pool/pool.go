package pool

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

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

type (
	ValidationFunction func(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool
	Metrics            map[string]float64
	MetricsTag         struct {
		Key   string
		Value string
	}
	// MetricsEmitterFunction the pool will simply emit raw metrics
	MetricsEmitterFunction func(metrics Metrics, tags []MetricsTag)
)

type AuroraPGPool struct {
	poolPtr                unsafe.Pointer
	queryValidationFunc    ValidationFunction
	logger                 *zap.Logger
	queryHealthCheckPeriod time.Duration
	closeChan              chan struct{}
	closeOnce              sync.Once
	config                 *Config
	metricsEmitter         MetricsEmitterFunction
}

func (p *AuroraPGPool) Close() {
	p.closeOnce.Do(func() {
		close(p.closeChan)
		p.getInnerPool().Close()
	})
}

func (p *AuroraPGPool) getInnerPool() *pgxpool.Pool {
	return (*pgxpool.Pool)(atomic.LoadPointer(&p.poolPtr))
}

func (p *AuroraPGPool) storeInnerPool(pool *pgxpool.Pool) {
	atomic.StorePointer(&p.poolPtr, unsafe.Pointer(pool))
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

func (p *AuroraPGPool) runValidator(parent context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool {
	tCtx, tCancel := context.WithTimeout(parent, time.Millisecond*500)
	defer tCancel()
	return p.queryValidationFunc(tCtx, conn, logger)
}

func (p *AuroraPGPool) checkQueryHealth() {
	stats := p.Stat()
	host := p.Config().ConnConfig.Host
	p.logger.Info("pool stats", zap.Int64("acquired", int64(stats.AcquiredConns())),
		zap.Int64("idle", int64(stats.IdleConns())),
		zap.Int64("max", int64(stats.MaxConns())))

	p.sendPoolConnMetrics(stats, host)

	ctx := context.Background()
	timedCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	conns := p.AcquireAllIdle(timedCtx)
	availableCount := len(conns)
	if availableCount == 0 { // TODO: Retry logic
		p.logger.Warn("Health check reported no available connections")
		return
	}

	p.logger.Info("started checkQueryHealth run..")
	destroyCount := 0
	if p.queryValidationFunc != nil {
		for _, conn := range conns {
			validated := p.runValidator(ctx, conn, p.logger)
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

	if availableCount > p.config.MinAvailableConnectionFailSize &&
		destroyCount > p.config.ValidationCountDestroyTrigger {
		p.logger.Sugar().Warnf("Destroying pool since > %d connections failed validation",
			p.config.ValidationCountDestroyTrigger)
		pool, err := pgxpool.ConnectConfig(ctx, p.getInnerPool().Config())
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
			tempPool := p.getInnerPool()
			p.storeInnerPool(pool) // set it to new before closing the retired pool
			p.logger.Info("Pool recreated")
			tempPool.Close() // close the old connections gracefully
			if p.metricsEmitter != nil {
				go p.metricsEmitter(
					Metrics{"pg_aurora_custom_db_destroy_count": 1},
					[]MetricsTag{{"pg_host", host}})
			}
		}
	}
	p.logger.Info("ended checkQueryHealth run..")
}

func (p *AuroraPGPool) sendPoolConnMetrics(stats *pgxpool.Stat, host string) {
	if p.metricsEmitter != nil {
		metrics := Metrics{
			"pg_aurora_custom_idle_conn":     float64(stats.IdleConns()),
			"pg_aurora_custom_acquired_conn": float64(stats.AcquiredConns()),
			"pg_aurora_custom_max_conn":      float64(stats.MaxConns()),
		}
		tags := []MetricsTag{{"pg_host", host}}
		p.metricsEmitter(metrics, tags)
	}
}

func (p *AuroraPGPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return p.getInnerPool().Acquire(ctx)
}

func (p *AuroraPGPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	return p.getInnerPool().AcquireFunc(ctx, f)
}

func (p *AuroraPGPool) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	return p.getInnerPool().AcquireAllIdle(ctx)
}

func (p *AuroraPGPool) Config() *pgxpool.Config {
	return p.getInnerPool().Config()
}

func (p *AuroraPGPool) Stat() *pgxpool.Stat {
	return p.getInnerPool().Stat()
}

func (p *AuroraPGPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return p.getInnerPool().Exec(ctx, sql, arguments...)
}

func (p *AuroraPGPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.getInnerPool().Query(ctx, sql, args...)
}

func (p *AuroraPGPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.getInnerPool().QueryRow(ctx, sql, args...)
}

func (p *AuroraPGPool) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{},
	f func(pgx.QueryFuncRow) error,
) (pgconn.CommandTag, error) {
	return p.getInnerPool().QueryFunc(ctx, sql, args, scans, f)
}

func (p *AuroraPGPool) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.getInnerPool().SendBatch(ctx, b)
}

func (p *AuroraPGPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.getInnerPool().Begin(ctx)
}

func (p *AuroraPGPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.getInnerPool().BeginTx(ctx, txOptions)
}

func (p *AuroraPGPool) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	return p.getInnerPool().BeginFunc(ctx, f)
}

func (p *AuroraPGPool) BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error {
	return p.getInnerPool().BeginTxFunc(ctx, txOptions, f)
}

func (p *AuroraPGPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string,
	rowSrc pgx.CopyFromSource,
) (int64, error) {
	return p.getInnerPool().CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (p *AuroraPGPool) Ping(ctx context.Context) error {
	return p.getInnerPool().Ping(ctx)
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
	if reflect.ValueOf(config.MinAvailableConnectionFailSize).IsZero() {
		config.MinAvailableConnectionFailSize = defaultMinAvailableConnectionFailSize
	}
	if reflect.ValueOf(config.ValidationCountDestroyTrigger).IsZero() {
		config.ValidationCountDestroyTrigger = defaultValidationCountDestroyTrigger
	}

	p := &AuroraPGPool{
		logger:                 logger,
		queryValidationFunc:    config.QueryValidator,
		queryHealthCheckPeriod: config.QueryHealthCheckPeriod,
		metricsEmitter:         config.MetricsEmitter,
		closeChan:              make(chan struct{}),
		config:                 config,
	}
	p.storeInnerPool(dbpool)
	// Start the validator
	if config.QueryValidator != nil {
		go p.backgroundQueryHealthCheck()
	}
	return p, nil
}
