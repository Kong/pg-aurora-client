package pool

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"runtime/debug"
	"time"
)

type Canary struct {
	ID          int64     `json:"ID"`
	LastUpdated time.Time `json:"lastUpdated"`
	DiffMS      float64   `json:"diffMS"`
}

var readerQuery = `SELECT id, ts, Extract(epoch FROM (current_timestamp - ts))*1000 AS diff_ms from canary;`

func reader(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool {
	var canary Canary
	var rows pgx.Rows
	var err error
	rows, err = conn.Query(ctx, readerQuery)
	if err != nil {
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
	logger.Info("Read Canary", zap.Int64("id", canary.ID), zap.Time("ts", canary.LastUpdated),
		zap.Float64("diff_ms", canary.DiffMS))

	return true
}

var DefaultReadValidator ValidationFunction = reader

var writerQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP RETURNING id,ts`

func writer(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool {
	var canary Canary
	var rows pgx.Rows
	var err error
	rows, err = conn.Query(ctx, writerQuery)
	if err != nil {
		return false
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&canary.ID,
			&canary.LastUpdated)
		if err != nil {
			logger.Sugar().Errorf("%s\n%s", err.Error(), debug.Stack())
			return false
		}
	}
	logger.Info("Write Canary", zap.Int64("id", canary.ID), zap.Time("ts", canary.LastUpdated))
	return true
}

var DefaultWriteValidator ValidationFunction = writer

var defaultQueryHealthCheckPeriod = time.Second * 60

type Config struct {
	QueryValidator         ValidationFunction
	QueryHealthCheckPeriod time.Duration
	PGXConfig              *pgxpool.Config
}
