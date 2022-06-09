package model

import (
	"database/sql"
	"time"
)

type ReplicaStatus struct {
	ServerID    string
	SessionID   string
	LastUpdated time.Time
}

var replicaStatusQuery = `SELECT SERVER_ID, SESSION_ID, LAST_UPDATE_TIMESTAMP FROM aurora_replica_status()
     WHERE EXTRACT(EPOCH FROM(NOW() - LAST_UPDATE_TIMESTAMP)) <= 300 OR SESSION_ID = 'MASTER_SESSION_ID'
     ORDER BY LAST_UPDATE_TIMESTAMP DESC`

type Store struct {
	DB *sql.DB
}

func (s *Store) GetReplicaStatus() ([]ReplicaStatus, error) {
	rsList := []ReplicaStatus{}
	rows, err := s.DB.Query(replicaStatusQuery)
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
