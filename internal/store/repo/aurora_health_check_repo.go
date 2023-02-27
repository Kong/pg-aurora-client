package repo

import (
	"context"
	"fmt"
	"github.com/kong/pg-aurora-client/internal/store/ent"
	"time"
)

type AuroraHealthCheckRepo struct {
	client *ent.Client
}

func NewAuroraHealthCheckRepo(client *ent.Client) *AuroraHealthCheckRepo {
	return &AuroraHealthCheckRepo{client: client}
}

func (repo AuroraHealthCheckRepo) Upsert(ctx context.Context, id *int) (*ent.AuroraHealthCheck, error) {
	var err error
	var entity *ent.AuroraHealthCheck
	if id == nil {
		entity, err = repo.client.AuroraHealthCheck.Create().SetID(
			1).SetTs(time.Now()).Save(ctx)
	} else {
		entity, err = repo.client.AuroraHealthCheck.UpdateOneID(*id).SetTs(time.Now()).Save(ctx)
	}
	return entity, err
}

func (repo AuroraHealthCheckRepo) Get(ctx context.Context, id *int) (*ent.AuroraHealthCheck, error) {
	var err error
	var entity *ent.AuroraHealthCheck
	if id == nil {
		return nil, fmt.Errorf("nil id passed")
	} else {
		entity, err = repo.client.AuroraHealthCheck.Get(ctx, *id)
	}
	return entity, err
}

func (repo AuroraHealthCheckRepo) Delete(ctx context.Context, id *int) (*ent.AuroraHealthCheck, error) {
	var err error
	var entity *ent.AuroraHealthCheck
	if id == nil {
		return nil, fmt.Errorf("nil id passed")
	} else {
		// Get it
		entity, err = repo.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		err = repo.client.AuroraHealthCheck.DeleteOneID(*id).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}
	return entity, err
}
