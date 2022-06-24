package model

import (
	"database/sql"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

type Store struct {
	DB     *sql.DB
	RODB   *sql.DB
	DBPool *pgxpool.Pool
	Logger *zap.Logger
}

type ReplicaStatus struct {
	ServerID    string    `json:"serverID"`
	SessionID   string    `json:"sessionID"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Foo struct {
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type ControlPlane struct {
	OrgID          string
	ControlPlaneID string
}
