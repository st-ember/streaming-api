package postgres_test

import (
	"testing"

	"github.com/st-ember/streaming-api/internal/adapter/driven/repo/postgres"
	"github.com/st-ember/streaming-api/internal/domain/user"
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
		_ = repo.SyncPermissions(t.Context(), slugs)
		err := repo.SyncPermissions(t.Context(), slugs)
		require.NoError(t, err)

		var count int
		err = TestDB.QueryRow("SELECT COUNT(*) FROM permissions").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("should do nothing on empty slugs slice", func(t *testing.T) {
		truncateAll(t)

		err := repo.SyncPermissions(t.Context(), []string{})
		require.NoError(t, err)

		var count int
		err = TestDB.QueryRow("SELECT COUNT(*) FROM permissions").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
}

func TestPostgresAuthRepo_SaveUserAndFindUserByKey(t *testing.T) {
	repo := postgres.NewPostgresAuthRepo(TestDB)

	t.Run("should save and find user by email", func(t *testing.T) {
		truncateAll(t)

		u, _ := user.NewUser("user-1", "test@test.com", "tester", "hashed")
		err := repo.SaveUser(t.Context(), u)
		require.NoError(t, err)

		// Find by Email
		found, err := repo.FindUserByKey(t.Context(), "test@test.com")
		require.NoError(t, err)
		require.Equal(t, u.ID, found.ID)
	})

	t.Run("should save and find user by username", func(t *testing.T) {
		truncateAll(t)

		u, _ := user.NewUser("user-1", "test@test.com", "tester", "hashed")
		err := repo.SaveUser(t.Context(), u)
		require.NoError(t, err)

		// Find by Email
		found, err := repo.FindUserByKey(t.Context(), "tester")
		require.NoError(t, err)
		require.Equal(t, u.ID, found.ID)

		// Case Insensitive Find
		found, err = repo.FindUserByKey(t.Context(), "TESTER")
		require.NoError(t, err)
		require.Equal(t, u.ID, found.ID)
	})

	t.Run("should fail when user does not exist", func(t *testing.T) {
		truncateAll(t)

		found, err := repo.FindUserByKey(t.Context(), "non-existent")
		require.Error(t, err)
		require.Nil(t, found)
	})
}

func TestPostgresAuthRepo_FindPermissionsByRole(t *testing.T) {
	repo := postgres.NewPostgresAuthRepo(TestDB)

	t.Run("should find permissions by role name", func(t *testing.T) {
		truncateAll(t)

		// Seed data
		_, _ = TestDB.Exec("INSERT INTO roles (id, name) VALUES ('r1', 'admin')")
		_, _ = TestDB.Exec("INSERT INTO permissions (id, slug) VALUES ('p1', 'video:delete'), ('p2', 'video:upload')")
		_, _ = TestDB.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ('r1', 'p1'), ('r1', 'p2')")

		perms, err := repo.FindPermissionsByRole(t.Context(), "admin")
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"video:delete", "video:upload"}, perms)
	})

	t.Run("should return empty slice for non-existent role", func(t *testing.T) {
		truncateAll(t)

		perms, err := repo.FindPermissionsByRole(t.Context(), "ghost")
		require.NoError(t, err)
		require.Empty(t, perms)
	})
}

func TestPostgresAuthRepo_SaveUserRole(t *testing.T) {
	repo := postgres.NewPostgresAuthRepo(TestDB)

	t.Run("should save user role and find permissions by user id", func(t *testing.T) {
		truncateAll(t)

		// Seed role and permissions
		_, _ = TestDB.Exec("INSERT INTO roles (id, name) VALUES ('r1', 'creator')")
		_, _ = TestDB.Exec("INSERT INTO permissions (id, slug) VALUES ('p1', 'video:upload')")
		_, _ = TestDB.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ('r1', 'p1')")

		// Save user
		u, _ := user.NewUser("user-1", "user@test.com", "user", "hash")
		_ = repo.SaveUser(t.Context(), u)

		// Assign role
		err := repo.SaveUserRole(t.Context(), u.ID, "creator")
		require.NoError(t, err)

		// Confirm user role was saved successfully
		perms, err := repo.FindPermissionsByUserID(t.Context(), u.ID)
		require.NoError(t, err)
		require.Contains(t, perms, "video:upload")
	})

	t.Run("should fail when saving for non-existent user", func(t *testing.T) {
		truncateAll(t)

		// Seed role and permissions
		_, _ = TestDB.Exec("INSERT INTO roles (id, name) VALUES ('r1', 'creator')")
		_, _ = TestDB.Exec("INSERT INTO permissions (id, slug) VALUES ('p1', 'video:upload')")
		_, _ = TestDB.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ('r1', 'p1')")

		// Assign role
		err := repo.SaveUserRole(t.Context(), "non-existent", "creator")
		require.ErrorContains(t, err, "23503")
	})

	t.Run("should fail when saving for non-existent role", func(t *testing.T) {
		truncateAll(t)

		// Seed role and permissions
		_, _ = TestDB.Exec("INSERT INTO permissions (id, slug) VALUES ('p1', 'video:upload')")
		_, _ = TestDB.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ('r1', 'p1')")

		// Save user
		u, _ := user.NewUser("user-1", "user@test.com", "user", "hash")
		_ = repo.SaveUser(t.Context(), u)

		// Assign role
		err := repo.SaveUserRole(t.Context(), u.ID, "non-existent")
		require.ErrorContains(t, err, "cannot find role id")
	})
}

func TestPostgresAuthRepo_FindPermissionsByUserID(t *testing.T) {
	repo := postgres.NewPostgresAuthRepo(TestDB)

	t.Run("should return empty for user with no role", func(t *testing.T) {
		truncateAll(t)

		u, _ := user.NewUser("user-1", "user@test.com", "user", "hash")
		_ = repo.SaveUser(t.Context(), u)

		perms, err := repo.FindPermissionsByUserID(t.Context(), u.ID)
		require.NoError(t, err)
		require.Empty(t, perms)
	})

	t.Run("should return empty for non-existent user id", func(t *testing.T) {
		truncateAll(t)

		perms, err := repo.FindPermissionsByUserID(t.Context(), "ghost")
		require.NoError(t, err)
		require.Empty(t, perms)
	})
}
