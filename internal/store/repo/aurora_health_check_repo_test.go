package repo

import (
	"context"
	"github.com/kong/pg-aurora-client/internal/store"
	"github.com/kong/pg-aurora-client/internal/store/ent"
	"github.com/kong/pg-aurora-client/pkg/entpgx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuroraHealthCheckRepo(t *testing.T) {
	// skip in the short mode
	if testing.Short() {
		return
	}
	// Setup database
	dbContainer, connPool, err := store.SetupTestDatabase()
	if err != nil {
		t.Error(err)
	}
	defer dbContainer.Terminate(context.Background())
	driver := entpgx.NewPgxPoolDriver(connPool)

	client := ent.NewClient(ent.Driver(driver))
	repo := NewAuroraHealthCheckRepo(client)
	// Run tests against db
	t.Run("FindExistingUserByUsername", func(t *testing.T) {
		upsert, err := repo.Upsert(context.Background(), nil)

		require.NoError(t, err)
		require.Equal(t, 1, upsert.ID)
		id := 1
		get, err := repo.Get(context.Background(), &id)
		require.NoError(t, err)
		require.Equal(t, 1, get.ID)
	})
}
