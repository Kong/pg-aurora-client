package model

import (
	"context"
	"errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kong/pg-aurora-client/pkg/metrics"
	"github.com/kong/pg-aurora-client/pkg/pool"
	"go.uber.org/zap"
	"sync"
	"time"
)

const defaultMaxConnections = 50
const defaultMinConnections = 20

var defaultLagCheckFrequency = time.Second * 60
var defaultBackoffInterval = time.Millisecond * 10 // keeping this low, otherwise it impacts least-count
var defaultLagReadRetries uint64 = 100             // fail after a second of retries

type Store struct {
	rwDBPool  pool.PGXConnPool
	roDBPool  pool.PGXConnPool
	Logger    *zap.Logger
	closeChan chan struct{}
	closeOnce sync.Once
}

func NewStore(logger *zap.Logger, pgc *PgConfig) (*Store, error) {
	dsn := getDSN(pgc)
	rodsn := getRODSN(pgc)

	rwPool, err := openPool(dsn, pgc, logger, pool.DefaultWriteValidator)
	if err != nil {
		return nil, err
	}
	logger.Info("established rw db connection to ", zap.String("host", rwPool.Config().ConnConfig.Host))
	roPool, err := openPool(rodsn, pgc, logger, pool.DefaultReadValidator)
	if err != nil {
		return nil, err
	}
	logger.Info("established ro db connection to ", zap.String("host", roPool.Config().ConnConfig.Host))

	store := &Store{
		rwDBPool:  rwPool,
		roDBPool:  roPool,
		Logger:    logger,
		closeChan: make(chan struct{}),
	}
	if store.rwDBPool != nil && store.roDBPool != nil {
		go store.backgroundLagCheck()
	}

	return store, nil
}

func (s *Store) Close() {
	s.closeOnce.Do(func() {
		close(s.closeChan)
		if s.rwDBPool != nil {
			s.rwDBPool.Close()
		}
		if s.roDBPool != nil {
			s.roDBPool.Close()
		}
	})
}

func (s *Store) backgroundLagCheck() {
	ticker := time.NewTicker(defaultLagCheckFrequency)
	defer ticker.Stop()
	for {
		select {
		case <-s.closeChan:
			s.Logger.Info("backgroundQueryHealthCheck exited..")
			return
		case <-ticker.C:
			s.checkReadLag()
		}
	}
}

func (s *Store) checkReadLag() {
	canary, err := s.UpdateReplicationCanary()
	if err != nil {
		s.Logger.Error("lag check update action error. Returning early", zap.Error(err))
		return
	}
	cb := backoff.NewConstantBackOff(defaultBackoffInterval)
	backoff.WithMaxRetries(cb, defaultLagReadRetries)
	err = backoff.Retry(func() error {
		canaryRead, err := s.GetReplicationCanary()
		if err != nil {
			s.Logger.Error("lag check read action error.", zap.Error(err))
			return err
		}
		if canary.ID != canaryRead.ID {
			s.Logger.Error("canary write and read are not the same.",
				zap.Int64("write ID", canary.ID),
				zap.Int64("read ID", canaryRead.ID))
			return errors.New("write ID not found during read")
		}
		go metrics.Gauge("pg_aurora_custom_replication_lag", canaryRead.DiffMS)
		s.Logger.Info("read lag measured", zap.Float64("duration_ms", canaryRead.DiffMS))
		return nil
	}, cb)
	if err != nil {
		s.Logger.Error("failed lag measurement", zap.Error(err))
	}
}

type ReplicaStatus struct {
	ServerID    string    `json:"serverID"`
	SessionID   string    `json:"sessionID"`
	LastUpdated time.Time `json:"lastUpdated"`
}

var replicaStatusQuery = `SELECT SERVER_ID, SESSION_ID, LAST_UPDATE_TIMESTAMP FROM aurora_replica_status()
     WHERE EXTRACT(EPOCH FROM(NOW() - LAST_UPDATE_TIMESTAMP)) <= 300 OR SESSION_ID = 'MASTER_SESSION_ID'
     ORDER BY LAST_UPDATE_TIMESTAMP DESC`

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

type PoolStats struct {
	AcquireCount    int64
	AcquireDuration time.Duration
	AcquiredConns   int32
	IdleConns       int32
	TotalConns      int32
	MaxConns        int32
}

func (s *Store) GetConnectionPoolStats(ro bool) *PoolStats {
	var stat *pgxpool.Stat
	if !ro {
		stat = s.rwDBPool.Stat()
	} else {
		stat = s.roDBPool.Stat()
	}
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

type Canary struct {
	ID          int64     `json:"ID"`
	LastUpdated time.Time `json:"lastUpdated"`
	DiffMS      float64   `json:"diffMS"`
}

var getCanaryQuery = `SELECT id, ts, Extract(epoch FROM (current_timestamp - ts))*1000 AS diff_ms from canary;`

var updateCanaryQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP RETURNING id,ts`

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

var getReplicationCanaryQuery = `SELECT id, ts, Extract(epoch FROM (current_timestamp - ts))*1000 AS diff_ms from 
                                 replication_canary;`

var updateReplicationCanaryQuery = `UPDATE replication_canary SET id=id +1, ts = CURRENT_TIMESTAMP RETURNING id,ts`

func (s *Store) UpdateReplicationCanary() (*Canary, error) {
	var canary Canary
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	rows, err = s.rwDBPool.Query(ctx, updateReplicationCanaryQuery)
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

func (s *Store) GetReplicationCanary() (*Canary, error) {
	var canary Canary
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	rows, err = s.roDBPool.Query(ctx, getReplicationCanaryQuery)
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
