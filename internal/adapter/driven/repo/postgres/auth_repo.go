package postgres

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"

	"database/sql"
)

type PostgresAuthRepo struct {
	db *sql.DB
}

func NewPostgresAuthRepo(db *sql.DB) repo.AuthRepo {
	return &PostgresAuthRepo{db}
}

func (ar *PostgresAuthRepo) SyncPermissions(ctx context.Context, slugs []string) error {
	query := `
		INSERT INTO permissions (id, slug)
		SELECT gen_random_uuid(), s
		FROM UNNEST($1::text[]) AS s
		ON CONFLICT (slug) DO NOTHING
		`

	_, err := ar.db.ExecContext(ctx, query, slugs)
	if err != nil {
		return fmt.Errorf("batch sync slugs: %w", err)
	}

	return nil
}
