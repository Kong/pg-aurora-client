package model

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

var getLastFooQuery = `SELECT ID, CREATED_AT, UPDATED_AT FROM FOO ORDER BY CREATED_AT DESC LIMIT 1`

var insertFoo = `INSERT INTO foo (id) VALUES (default)`

type Foo struct {
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
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
