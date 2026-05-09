package postgres_test

import (
	"testing"

	"github.com/st-ember/streaming-api/internal/adapter/driven/repo/postgres"
	"github.com/stretchr/testify/require"
)

func TestPostgresAuthRepo_SyncPermissions(t *testing.T) {
	repo := postgres.NewPostgresAuthRepo(TestDB)

	t.Run("should insert all permissions on first sync", func(t *testing.T) {
		truncateAll(t)

		slugs := []string{"video:upload", "video:delete"}
		err := repo.SyncPermissions(t.Context(), slugs)
		require.NoError(t, err)

		var count int
		err = TestDB.QueryRow("SELECT COUNT(*) FROM permissions").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})

	t.Run("should be idempotent and not fail on duplicate sync", func(t *testing.T) {
		truncateAll(t)

		slugs := []string{"video:upload"}

		// Run twice
		err := repo.SyncPermissions(t.Context(), slugs)
		require.NoError(t, err)
		err = repo.SyncPermissions(t.Context(), slugs)
		require.NoError(t, err)

		var count int
		err = TestDB.QueryRow("SELECT COUNT(*) FROM permissions").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("should add new permissions to existing ones", func(t *testing.T) {
		truncateAll(t)

		// Initial sync
		_ = repo.SyncPermissions(t.Context(), []string{"video:upload"})

		// Sync with additional permission
		slugs := []string{"video:upload", "video:archive"}
		err := repo.SyncPermissions(t.Context(), slugs)
		require.NoError(t, err)

		var count int
		err = TestDB.QueryRow("SELECT COUNT(*) FROM permissions").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})
}
