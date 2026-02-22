package postgres

import (
	"database/sql"
	"log"
	"os"
	"testing"

	embeddedpgx "github.com/fergusstrange/embedded-postgres"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

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
	testDB, err = sql.Open("pgx", connString)
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
	`
	_, err = testDB.Exec(createTablesSQL)
	if err != nil {
		log.Fatalf("create tables in embdedded postgres: %v", err)
	}

	// Run all the tests in the package
	code := m.Run()

	// Stop embdedded postgres
	if err := postgres.Stop(); err != nil {
		log.Fatalf("stop embdedded postgres: %v", err)
	}

	if err := testDB.Close(); err != nil {
		log.Fatalf("close test db connection: %v", err)
	}

	// Exit with the tests' exit code
	os.Exit(code)
}

func beginTx(t *testing.T) *sql.Tx {
	tx, err := testDB.BeginTx(t.Context(), nil)
	require.NoError(t, err)

	_, err = tx.ExecContext(t.Context(), "TRUNCATE videos, jobs RESTART IDENTITY CASCADE;")
	require.NoError(t, err)

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
