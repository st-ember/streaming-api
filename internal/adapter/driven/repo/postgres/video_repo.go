package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/st-ember/streaming-api/internal/domain/video"
)

type PostgresVideoRepo struct {
	tx *sql.Tx
}

func NewPostgresVideoRepo(tx *sql.Tx) *PostgresVideoRepo {
	return &PostgresVideoRepo{tx}
}

// Save upserts the specified video
func (r *PostgresVideoRepo) Save(ctx context.Context, video *video.Video) error {
	query := `
		INSERT INTO videos (id, title, description, duration, filename, 
		resource_id, status, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		duration = EXCLUDED.duration,
		filename = EXCLUDED.filename,
		resource_id = EXCLUDED.resource_id,
		status = EXCLUDED.status,
		updated_at = EXCLUDED.updated_at;
	`

	_, err := r.tx.ExecContext(ctx, query,
		video.ID, video.Title, video.Description, video.Duration,
		video.Filename, video.ResourceID, video.Status, video.CreatedAt, video.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save video %s: %w", video.ID, err)
	}

	return nil
}

// FindByID finds the video entity specified by the id param
func (r *PostgresVideoRepo) FindByID(ctx context.Context, id string) (*video.Video, error) {
	v := &video.Video{}

	query := `
		SELECT id, title, description, duration, filename,
		resource_id, status, created_at, updated_at
		FROM videos
		WHERE id = $1;
	`

	err := r.tx.QueryRowContext(ctx, query, id).Scan(
		&v.ID,
		&v.Title,
		&v.Description,
		&v.Duration,
		&v.Filename,
		&v.ResourceID,
		&v.Status,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("scan video %s data: %w", id, err)
	}

	return v, nil
}
