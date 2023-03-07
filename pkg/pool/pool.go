package pool

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"time"
)

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
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Ping(ctx context.Context) error
	Reset()
}

type (
	ValidationFunction func(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool
	Metric             struct {
		Key   string
		Value float64
	}
	MetricsTag struct {
		Key   string
		Value string
	}
	// MetricsEmitterFunction the pool can emit the pgxpool.Stat or raw metrics
	MetricsEmitterFunction func(metrics interface{}, tags []MetricsTag)
)

type AuroraPGPool struct {
	innerPool                      *pgxpool.Pool
	queryValidationFunc            ValidationFunction
	logger                         *zap.Logger
	queryHealthCheckPeriod         time.Duration
	closeChan                      chan struct{}
	closeOnce                      sync.Once
	metricsEmitter                 MetricsEmitterFunction
	minAvailableConnectionFailSize int
	validationCountDestroyTrigger  int
	queryValidationTimeout         time.Duration
}

func (p *AuroraPGPool) Close() {
	p.closeOnce.Do(func() {
		close(p.closeChan)
		p.innerPool.Close()
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

func (p *AuroraPGPool) runValidator(parent context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool {
	tCtx, tCancel := context.WithTimeout(parent, p.queryValidationTimeout)
	defer tCancel()
	return p.queryValidationFunc(tCtx, conn, logger)
}

func (p *AuroraPGPool) checkQueryHealth() {
	stats := p.Stat()
	host := p.Config().ConnConfig.Host
	p.logger.Info("pool stats", zap.Int64("acquired", int64(stats.AcquiredConns())),
		zap.Int64("idle", int64(stats.IdleConns())),
		zap.Int64("max", int64(stats.MaxConns())))

	if p.metricsEmitter != nil {
		tags := []MetricsTag{{"pg_host", host}}
		p.metricsEmitter(stats, tags)
	}

	ctx := context.Background()
	timedCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	conns := p.AcquireAllIdle(timedCtx)
	availableCount := len(conns)
	if availableCount == 0 { // TODO: Retry logic
		p.logger.Warn("Health check reported no available connections")
		return
	}

	p.logger.Debug("started checkQueryHealth run..")
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
	p.logger.Info("Connections pool state", zap.String("pg_host", host),
		zap.Int("availableCount:", availableCount), zap.Int("destroyed", destroyCount))

	if availableCount > p.minAvailableConnectionFailSize &&
		destroyCount > p.validationCountDestroyTrigger {
		p.logger.Sugar().Warnf("Resetting pool since > %d connections failed validation",
			p.validationCountDestroyTrigger)
		p.innerPool.Reset()
		p.logger.Info("Pool reset complete")
		if p.metricsEmitter != nil {
			go p.metricsEmitter(
				Metric{"pg_aurora_custom_db_destroy_count", 1},
				[]MetricsTag{{"pg_host", host}})
		}
	}
	p.logger.Debug("ended checkQueryHealth run..")
}

func (p *AuroraPGPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return p.innerPool.Acquire(ctx)
}

func (p *AuroraPGPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	return p.innerPool.AcquireFunc(ctx, f)
}

func (p *AuroraPGPool) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	return p.innerPool.AcquireAllIdle(ctx)
}

func (p *AuroraPGPool) Config() *pgxpool.Config {
	return p.innerPool.Config()
}

func (p *AuroraPGPool) Stat() *pgxpool.Stat {
	return p.innerPool.Stat()
}

func (p *AuroraPGPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return p.innerPool.Exec(ctx, sql, arguments...)
}

func (p *AuroraPGPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.innerPool.Query(ctx, sql, args...)
}

func (p *AuroraPGPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.innerPool.QueryRow(ctx, sql, args...)
}

func (p *AuroraPGPool) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.innerPool.SendBatch(ctx, b)
}

func (p *AuroraPGPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.innerPool.Begin(ctx)
}

func (p *AuroraPGPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.innerPool.BeginTx(ctx, txOptions)
}

func (p *AuroraPGPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string,
	rowSrc pgx.CopyFromSource,
) (int64, error) {
	return p.innerPool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (p *AuroraPGPool) Ping(ctx context.Context) error {
	return p.innerPool.Ping(ctx)
}

func (p *AuroraPGPool) Reset() {
	p.innerPool.Reset()
}

func NewAuroraPool(ctx context.Context, config *Config, logger *zap.Logger) (*AuroraPGPool, error) {
	// Intentionally not being aggressive since we have 2 background check threads
	config.PGXConfig.HealthCheckPeriod = time.Minute * 5
	dbpool, err := pgxpool.NewWithConfig(ctx, config.PGXConfig)
	if err != nil {
		return nil, err
	}
	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	queryValidationTimeout := config.QueryValidationTimeout
	queryHealthCheckPeriod := config.QueryHealthCheckPeriod
	minAvailableConnectionFailSize := config.MinAvailableConnectionFailSize
	validationCountDestroyTrigger := config.ValidationCountDestroyTrigger

	if reflect.ValueOf(config.QueryValidationTimeout).IsZero() {
		queryValidationTimeout = defaultQueryValidationTimeout
	}
	if reflect.ValueOf(config.QueryHealthCheckPeriod).IsZero() {
		queryHealthCheckPeriod = defaultQueryHealthCheckPeriod
	}
	if reflect.ValueOf(config.MinAvailableConnectionFailSize).IsZero() {
		minAvailableConnectionFailSize = defaultMinAvailableConnectionFailSize
	}
	if reflect.ValueOf(config.ValidationCountDestroyTrigger).IsZero() {
		validationCountDestroyTrigger = defaultValidationCountDestroyTrigger
	}

	p := &AuroraPGPool{
		logger:                         logger,
		queryValidationFunc:            config.QueryValidator,
		queryHealthCheckPeriod:         queryHealthCheckPeriod,
		metricsEmitter:                 config.MetricsEmitter,
		queryValidationTimeout:         queryValidationTimeout,
		minAvailableConnectionFailSize: minAvailableConnectionFailSize,
		validationCountDestroyTrigger:  validationCountDestroyTrigger,
		closeChan:                      make(chan struct{}),
	}
	p.innerPool = dbpool
	// Start the validator
	if config.QueryValidator != nil {
		go p.backgroundQueryHealthCheck()
	}
	return p, nil
}
