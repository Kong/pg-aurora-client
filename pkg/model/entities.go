package model

import (
	"context"
	"database/sql"
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
	var rows *sql.Rows
	var err error
	if ro && s.RODB != nil {
		rows, err = s.RODB.Query(replicaStatusQuery)
	} else {
		if ro {
			s.Logger.Warn("using rw connection because there ro connection is not injected")
		}
		rows, err = s.DB.Query(replicaStatusQuery)
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
	var rows *sql.Rows
	var err error
	if s.RODB != nil {
		rows, err = s.RODB.Query(getLastFooQuery)
	} else {
		rows, err = s.DB.Query(getLastFooQuery)
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
	exec, err := s.DB.Exec(insertFoo)
	var affected int64
	if err != nil {
		return affected, err
	}
	affected, err = exec.RowsAffected()
	if err != nil {
		return affected, err
	}
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
