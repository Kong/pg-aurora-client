package model

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/kong/pg-aurora-client/pkg/pool"
	"go.uber.org/zap"
	"time"
)

const defaultMaxConnections = 50
const defaultMinConnections = 20

type ReplicaStatus struct {
	ServerID    string    `json:"serverID"`
	SessionID   string    `json:"sessionID"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Canary struct {
	ID          int64     `json:"ID"`
	LastUpdated time.Time `json:"lastUpdated"`
	DiffMS      float64   `json:"diffMS"`
}

type Foo struct {
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

var replicaStatusQuery = `SELECT SERVER_ID, SESSION_ID, LAST_UPDATE_TIMESTAMP FROM aurora_replica_status()
     WHERE EXTRACT(EPOCH FROM(NOW() - LAST_UPDATE_TIMESTAMP)) <= 300 OR SESSION_ID = 'MASTER_SESSION_ID'
     ORDER BY LAST_UPDATE_TIMESTAMP DESC`

var getLastFooQuery = `SELECT ID, CREATED_AT, UPDATED_AT FROM FOO ORDER BY CREATED_AT DESC LIMIT 1`

var insertFoo = `INSERT INTO foo (id) VALUES (default)`

var getCanaryQuery = `SELECT id, ts, Extract(epoch FROM (current_timestamp - ts))*1000 AS diff_ms from canary;`

var updateCanaryQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP RETURNING id,ts`

type Store struct {
	rwDBPool pool.PGXConnPool
	roDBPool pool.PGXConnPool
	Logger   *zap.Logger
}

func NewStore(logger *zap.Logger, pgc *PgConfig) (*Store, error) {

	dsn := getDSN(pgc)
	rodsn := getRODSN(pgc)

	rwPool, err := openPool(dsn, pgc, logger, pool.DefaultWriteValidator)
	if err != nil {
		return nil, err
	}
	logger.Info("Established DB Connection")
	roPool, err := openPool(rodsn, pgc, logger, pool.DefaultReadValidator)
	if err != nil {
		return nil, err
	}
	logger.Info("Established RO DB Connection")

	store := &Store{
		rwDBPool: rwPool,
		roDBPool: roPool,
		Logger:   logger,
	}
	return store, nil
}

type PoolStats struct {
	AcquireCount    int64
	AcquireDuration time.Duration
	AcquiredConns   int32
	IdleConns       int32
	TotalConns      int32
	MaxConns        int32
}

func (s *Store) GetReplicaStatus(ro bool) ([]ReplicaStatus, error) {
	rsList := []ReplicaStatus{}
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	if ro && s.roDBPool != nil {
		rows, err = s.roDBPool.Query(ctx, replicaStatusQuery)
	} else {
		if ro {
			s.Logger.Warn("using rw connection because there ro connection is not injected")
		}
		rows, err = s.rwDBPool.Query(ctx, replicaStatusQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rs ReplicaStatus
		err := rows.Scan(
			&rs.ServerID,
			&rs.SessionID,
			&rs.LastUpdated)
		if err != nil {
			return nil, err
		}
		rsList = append(rsList, rs)
	}
	return rsList, nil
}

func (s *Store) GetMostRecentFoo() (*Foo, error) {
	var foo Foo
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	if s.roDBPool != nil {
		rows, err = s.roDBPool.Query(ctx, getLastFooQuery)
	} else {
		rows, err = s.rwDBPool.Query(ctx, getLastFooQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&foo.ID,
			&foo.Created,
			&foo.LastUpdated)
		if err != nil {
			return nil, err
		}
	}
	return &foo, nil
}

func (s *Store) InsertFoo() (int64, error) {
	ctx := context.Background()
	exec, err := s.rwDBPool.Exec(ctx, insertFoo)
	var affected int64
	if err != nil {
		return affected, err
	}
	affected = exec.RowsAffected()
	return affected, nil
}

func (s *Store) GetConnectionPoolStats() *PoolStats {
	stat := s.rwDBPool.Stat()
	poolstats := &PoolStats{
		AcquireCount:    stat.AcquireCount(),
		AcquireDuration: stat.AcquireDuration(),
		AcquiredConns:   stat.AcquiredConns(),
		IdleConns:       stat.IdleConns(),
		TotalConns:      stat.TotalConns(),
		MaxConns:        stat.MaxConns(),
	}
	return poolstats
}

func (s *Store) GetROConnectionPoolStats() *PoolStats {
	stat := s.roDBPool.Stat()
	poolstats := &PoolStats{
		AcquireCount:    stat.AcquireCount(),
		AcquireDuration: stat.AcquireDuration(),
		AcquiredConns:   stat.AcquiredConns(),
		IdleConns:       stat.IdleConns(),
		TotalConns:      stat.TotalConns(),
		MaxConns:        stat.MaxConns(),
	}
	return poolstats
}

func (s *Store) UpdateCanary() (*Canary, error) {
	var canary Canary
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	rows, err = s.rwDBPool.Query(ctx, updateCanaryQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&canary.ID,
			&canary.LastUpdated)
		if err != nil {
			return nil, err
		}
	}
	return &canary, nil
}

func (s *Store) GetCanary() (*Canary, error) {
	var canary Canary
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	rows, err = s.roDBPool.Query(ctx, getCanaryQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&canary.ID,
			&canary.LastUpdated,
			&canary.DiffMS)
		if err != nil {
			return nil, err
		}
	}
	return &canary, nil
}

func (s *Store) Close() {
	if s.rwDBPool != nil {
		s.rwDBPool.Close()
	}

	if s.roDBPool != nil {
		s.roDBPool.Close()
	}
}
