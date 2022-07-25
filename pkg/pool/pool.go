package pool

import (
	"context"
	"github.com/kong/pg-aurora-client/pkg/model"
)

type ValidateWriteFunc func(ctx context.Context, store *model.Store) (bool, error)

type ValidateReadFunc func(ctx context.Context, store *model.Store) (bool, error)

type auroraValidator struct {
	store             *model.Store
	validateWriteFunc ValidateWriteFunc
	validateReadFunc  ValidateReadFunc
}

type ConnectionValidator interface {
	ValidateWrite(ctx context.Context) (bool, error)
	ValidateRead(ctx context.Context) (bool, error)
}

func (v *auroraValidator) ValidateWrite(ctx context.Context) (bool, error) {
	return v.validateWriteFunc(ctx, v.store)
}

func (v *auroraValidator) ValidateRead(ctx context.Context) (bool, error) {
	return v.validateReadFunc(ctx, v.store)
}

func NewDefaultConnValidator(store *model.Store, writeFunc ValidateWriteFunc, readFunc ValidateReadFunc) ConnectionValidator {
	return &auroraValidator{store: store, validateWriteFunc: writeFunc, validateReadFunc: readFunc}
}
