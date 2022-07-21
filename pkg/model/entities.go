package model

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

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

var updateHealthQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP`

var roHealthQuery = `SELECT id, ts, Extract(epoch FROM (current_timestamp - ts))*1000 AS diff_ms from canary;`

type Store struct {
	DBPool   *pgxpool.Pool
	RODBPool *pgxpool.Pool
	Logger   *zap.Logger
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
	if ro && s.RODBPool != nil {
		rows, err = s.RODBPool.Query(ctx, replicaStatusQuery)
	} else {
		if ro {
			s.Logger.Warn("using rw connection because there ro connection is not injected")
		}
		rows, err = s.DBPool.Query(ctx, replicaStatusQuery)
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
	if s.RODBPool != nil {
		rows, err = s.RODBPool.Query(ctx, getLastFooQuery)
	} else {
		rows, err = s.DBPool.Query(ctx, getLastFooQuery)
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
	exec, err := s.DBPool.Exec(ctx, insertFoo)
	var affected int64
	if err != nil {
		return affected, err
	}
	affected = exec.RowsAffected()
	return affected, nil
}

func (s *Store) GetConnectionPoolStats() *PoolStats {
	stat := s.DBPool.Stat()
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
	stat := s.RODBPool.Stat()
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

func (s *Store) UpdatePoolHealthCheck() (int64, error) {
	ctx := context.Background()
	exec, err := s.DBPool.Exec(ctx, updateHealthQuery)
	var affected int64
	if err != nil {
		return affected, err
	}
	affected = exec.RowsAffected()
	return affected, nil
}

func (s *Store) GetPoolHealthCheck() (*Canary, error) {
	var canary Canary
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	if s.RODBPool != nil {
		rows, err = s.RODBPool.Query(ctx, roHealthQuery)
	} else {
		rows, err = s.DBPool.Query(ctx, roHealthQuery)
	}
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
