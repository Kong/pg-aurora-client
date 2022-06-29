package model

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

type Store struct {
	DBPool   *pgxpool.Pool
	RODBPool *pgxpool.Pool
	Logger   *zap.Logger
}

type ReplicaStatus struct {
	ServerID    string    `json:"serverID"`
	SessionID   string    `json:"sessionID"`
	LastUpdated time.Time `json:"lastUpdated"`
	Host        string    `json:"host"`
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

type PoolStats struct {
	AcquireCount    int64
	AcquireDuration time.Duration
	AcquiredConns   int32
	IdleConns       int32
	TotalConns      int32
	MaxConns        int32
}
