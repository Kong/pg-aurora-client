package pool

import (
	"context"
	"github.com/kong/pg-aurora-client/pkg/model"
)

func testWriterConnValidator(ctx context.Context, store *model.Store) (bool, error) {
	return false, nil
}

func testReaderConnValidator(ctx context.Context, store *model.Store) (bool, error) {
	return true, nil
}
