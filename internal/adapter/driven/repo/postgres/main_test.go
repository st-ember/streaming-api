package postgres_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	embeddedpgx "github.com/fergusstrange/embedded-postgres"
	"github.com/stretchr/testify/require"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	config := embeddedpgx.DefaultConfig().
		Port(5433).
		Logger(nil)

	// Create embdedded postgres
	postgres := embeddedpgx.NewDatabase(config)
	if err := postgres.Start(); err != nil {
		log.Fatalf("start embdedded postgres: %v", err)
	}

	// Connect to embedded postgres
	connString := "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
	var err error
	TestDB, err = sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("connect to embdedded postgres: %v", err)
	}

	// Create database schema
	createTablesSQL := `
        CREATE TABLE IF NOT EXISTS videos (
            id TEXT PRIMARY KEY, title TEXT, description TEXT, duration BIGINT,
            filename TEXT, resource_id TEXT, status TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
        );
        CREATE TABLE IF NOT EXISTS jobs (
           id TEXT PRIMARY KEY, video_id TEXT, type TEXT, status TEXT,
           result TEXT, error_msg TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
        );

        -- RBAC Tables
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL,
            created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
        );
        CREATE TABLE IF NOT EXISTS roles (
            id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL
        );
        CREATE TABLE IF NOT EXISTS permissions (
            id TEXT PRIMARY KEY, slug TEXT UNIQUE NOT NULL
        );
        CREATE TABLE IF NOT EXISTS user_roles (
            user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
            role_id TEXT REFERENCES roles(id) ON DELETE CASCADE,
            PRIMARY KEY (user_id, role_id)
        );
        CREATE TABLE IF NOT EXISTS role_permissions (
            role_id TEXT REFERENCES roles(id) ON DELETE CASCADE,
            permission_id TEXT REFERENCES permissions(id) ON DELETE CASCADE,
            PRIMARY KEY (role_id, permission_id)
        );
	`
	_, err = TestDB.Exec(createTablesSQL)
	if err != nil {
		log.Fatalf("create tables in embdedded postgres: %v", err)
	}

	// Run all the tests in the package
	code := m.Run()

	// Stop embdedded postgres
	if err := postgres.Stop(); err != nil {
		log.Fatalf("stop embdedded postgres: %v", err)
	}

	if err := TestDB.Close(); err != nil {
		log.Fatalf("close test db connection: %v", err)
	}

	// Exit with the tests' exit code
	os.Exit(code)
}

func beginTx(t *testing.T) *sql.Tx {
	tx, err := TestDB.BeginTx(t.Context(), nil)
	require.NoError(t, err)

	_, err = tx.ExecContext(t.Context(), "TRUNCATE videos, jobs, users, roles, permissions, user_roles, role_permissions RESTART IDENTITY CASCADE;")
	require.NoError(t, err)

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}

func truncateAll(t *testing.T) {
	_, err := TestDB.ExecContext(t.Context(), "TRUNCATE videos, jobs, users, roles, permissions, user_roles, role_permissions RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}
