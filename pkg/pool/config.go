package pool

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Canary struct {
	ID          int64     `json:"ID"`
	LastUpdated time.Time `json:"lastUpdated"`
	DiffMS      float64   `json:"diffMS"`
}

var readerQuery = `SELECT id, ts, Extract(epoch FROM (current_timestamp - ts))*1000 AS diff_ms from canary;`

func readerValidator(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool {
	var canary Canary
	var rows pgx.Rows
	var err error
	rows, err = conn.Query(ctx, readerQuery)
	if err != nil {
		logger.Error("read validation failed", zap.Error(err))
		return false
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&canary.ID,
			&canary.LastUpdated,
			&canary.DiffMS)
		if err != nil {
			logger.Sugar().Errorf("%s\n%s", err.Error(), debug.Stack())
			return false
		}
	}
	logger.Info("healthcheck read canary", zap.Int64("id", canary.ID), zap.Time("ts", canary.LastUpdated),
		zap.Float64("diff_ms", canary.DiffMS))
	return true
}

var DefaultReaderValidator ValidationFunction = readerValidator

var writeQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP`

func writeValidator(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool {
	exec, err := conn.Exec(ctx, writeQuery)
	if err != nil {
		logger.Error("write validation failed", zap.Error(err))
		return false
	}

	logger.Info("healthcheck write canary", zap.Int64("rowsUpdated", exec.RowsAffected()))
	return true
}

var DefaultWriteValidator ValidationFunction = writeValidator

const (
	// DefaultQueryHealthCheckPeriod is the default value for Config.QueryHealthCheckPeriod.
	DefaultQueryHealthCheckPeriod = time.Second * 60

	// DefaultMinAvailableConnectionFailSize is the default value for Config.MinAvailableConnectionFailSize.
	DefaultMinAvailableConnectionFailSize = 3

	// DefaultValidationCountDestroyTrigger is the default value for Config.ValidationCountDestroyTrigger.
	DefaultValidationCountDestroyTrigger = 2

	// DefaultQueryValidationTimeout is the default value for Config.QueryValidationTimeout.
	DefaultQueryValidationTimeout = time.Millisecond * 500

	// DefaultPGXHealthCheckPeriod is used to limit PGX's own internal health check
	// period, as when AuroraPGPool is used, there are two background check threads.
	DefaultPGXHealthCheckPeriod = time.Minute * 5
)

// Config is used to instantiate a new AuroraPGPool.
type Config struct {
	// QueryValidator represents the required health check validation function.
	QueryValidator ValidationFunction

	// QueryValidationTimeout represents how long the query validation function is allowed to run.
	// Defaulted to DefaultQueryValidationTimeout when not specified.
	QueryValidationTimeout time.Duration

	// QueryHealthCheckPeriod represents how often the provided Config.QueryValidator function will run.
	// Defaulted to DefaultQueryHealthCheckPeriod when not specified.
	QueryHealthCheckPeriod time.Duration

	// MinAvailableConnectionFailSize is used in conjunction with Config.ValidationCountDestroyTrigger, and gates
	// when all connections on the pool are allowed to be reset. Specifically, the number of active connections
	// at the time of validation must be larger than this value, in order for all connections to be reset.
	//
	// Defaulted to DefaultMinAvailableConnectionFailSize when not specified.
	//
	// TODO(tjasko): This behavior seems strange and documentation is not provided on why this
	//  was done. Leaving this for a rainy day to figure out if this needs to be kept.
	MinAvailableConnectionFailSize int

	// ValidationCountDestroyTrigger represents how many consecutive validation attempts need to fail until all
	// connections on the pool are reset. When this count is reached, the pool will be reset on the next attempt.
	//
	// Defaulted to DefaultValidationCountDestroyTrigger when not specified.
	ValidationCountDestroyTrigger int

	// MetricsEmitter is an optional function used to collect metrics.
	MetricsEmitter MetricsEmitterFunction

	// PGXConfig is used to instantiate a new PGX pool instance on
	// behalf of the caller. Must not be used with Config.PGXPool.
	PGXConfig *pgxpool.Config

	// PGXPool is used to pass in a pre-instantiated PGX pool.
	// Must not be used with Config.PGXConfig.
	//
	// The caller is expected to set PGX's health check period to
	// an appropriate value, e.g.: DefaultPGXHealthCheckPeriod.
	PGXPool *pgxpool.Pool
}

func (c *Config) validate() error {
	c.setDefaults()

	if (c.PGXPool != nil && c.PGXConfig != nil) || (c.PGXPool == nil && c.PGXConfig == nil) {
		return errors.New("must specify a PGX pool instance or config to instantiate one, not both")
	}

	if c.QueryValidator == nil {
		// TODO(tjasko): The default query validation function really should just be executing a
		//  `SHOW transaction_read_only` SQL query, as writing to a table is unnecessary & wasteful.
		//
		// This would be similar to what AWS's Java Aurora DB driver does:
		// https://github.com/awslabs/aws-advanced-jdbc-wrapper
		return errors.New("must specify a query validator function")
	}

	return nil
}

func (c *Config) setDefaults() {
	if c.MinAvailableConnectionFailSize == 0 {
		c.MinAvailableConnectionFailSize = DefaultMinAvailableConnectionFailSize
	}
	if c.QueryHealthCheckPeriod == 0 {
		c.QueryHealthCheckPeriod = DefaultQueryHealthCheckPeriod
	}
	if c.QueryValidationTimeout == 0 {
		c.QueryValidationTimeout = DefaultQueryValidationTimeout
	}
	if c.ValidationCountDestroyTrigger == 0 {
		c.ValidationCountDestroyTrigger = DefaultValidationCountDestroyTrigger
	}
	if c.PGXConfig != nil {
		c.PGXConfig.HealthCheckPeriod = DefaultPGXHealthCheckPeriod
	}
}
