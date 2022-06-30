package model

import (
	"context"
	"github.com/jackc/pgx/v4"
)

var replicaStatusQuery = `SELECT SERVER_ID, SESSION_ID, LAST_UPDATE_TIMESTAMP FROM aurora_replica_status()
     WHERE EXTRACT(EPOCH FROM(NOW() - LAST_UPDATE_TIMESTAMP)) <= 300 OR SESSION_ID = 'MASTER_SESSION_ID'
     ORDER BY LAST_UPDATE_TIMESTAMP DESC`

var getLastFooQuery = `SELECT ID, CREATED_AT, UPDATED_AT FROM FOO ORDER BY CREATED_AT DESC LIMIT 1`

var insertFoo = `INSERT INTO foo (id) VALUES (default)`

var getControlPlanesQuery = `SELECT org_id, control_plane_id FROM default_runtime_group_relations`

func (s *Store) GetReplicaStatus(ro bool) ([]ReplicaStatus, error) {
	rsList := []ReplicaStatus{}
	var rows pgx.Rows
	var err error
	var host string
	ctx := context.Background()
	if ro && s.RODBPool != nil {
		rows, err = s.RODBPool.Query(ctx, replicaStatusQuery)
		host = s.RODBPool.Config().ConnConfig.Host
	} else {
		host = s.DBPool.Config().ConnConfig.Host
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
		rs.Host = host
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

func (s *Store) GetControlPlanes() ([]ControlPlane, error) {
	cpList := []ControlPlane{}
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	rows, err = s.DBPool.Query(ctx, getControlPlanesQuery)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cp ControlPlane
		err := rows.Scan(
			&cp.OrgID,
			&cp.ControlPlaneID)
		if err != nil {
			return nil, err
		}
		cpList = append(cpList, cp)
	}
	return cpList, nil
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
