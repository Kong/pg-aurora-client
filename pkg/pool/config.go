package pool

import (
	"context"
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

var (
	defaultQueryHealthCheckPeriod         = time.Second * 60
	defaultMinAvailableConnectionFailSize = 3
	defaultValidationCountDestroyTrigger  = 2
	defaultQueryValidationTimeout         = time.Millisecond * 500
)

type Config struct {
	QueryValidator                 ValidationFunction
	QueryValidationTimeout         time.Duration
	QueryHealthCheckPeriod         time.Duration
	PGXConfig                      *pgxpool.Config
	MinAvailableConnectionFailSize int
	ValidationCountDestroyTrigger  int
	MetricsEmitter                 MetricsEmitterFunction
}
