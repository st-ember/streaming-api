package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/st-ember/streaming-api/internal/domain/job"
)

type PostgresJobRepo struct {
	tx *sql.Tx
}

func NewPostgresJobRepo(tx *sql.Tx) *PostgresJobRepo {
	return &PostgresJobRepo{tx}
}

// Save upserts the specified job
func (r *PostgresJobRepo) Save(ctx context.Context, job *job.Job) error {
	query := `
		INSERT INTO jobs (id, video_id, type, status, 
		result, error_msg, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
		status = EXCLUDED.status,
		result = EXCLUDED.result,
		error_msg = EXCLUDED.error_msg,
		updated_at = EXCLUDED.updated_at;
	`

	_, err := r.tx.ExecContext(ctx, query,
		job.ID, job.VideoID, job.Type, job.Status, job.Result,
		job.ErrorMsg, job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save job %s: %w", job.ID, err)
	}

	return nil
}

func (r *PostgresJobRepo) FindByVideoID(ctx context.Context, id string) (*job.Job, error) {
	j := &job.Job{}

	query := `
		SELECT id, video_id, type, status, result, error_msg, created_at, updated_at
		FROM jobs
		WHERE video_id = $1
		ORDER BY created_at DESC
		LIMIT 1;
	`

	err := r.tx.QueryRowContext(ctx, query, id).Scan(
		&j.ID,
		&j.VideoID,
		&j.Type,
		&j.Status,
		&j.Result,
		&j.ErrorMsg,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("scan job data: %w", err)
	}

	return j, nil
}

// FindNextPendingTranscodeJob finds the oldest job that can be transcoded
func (r *PostgresJobRepo) FindNextPendingTranscodeJob(ctx context.Context) (*job.Job, error) {
	j := &job.Job{}

	query := `
		SELECT id, video_id, type, status, created_at, updated_at
		FROM jobs
		WHERE status = 'pending' AND type = 'transcode'
		ORDER BY created_at
		LIMIT 1;
	`

	err := r.tx.QueryRowContext(ctx, query).Scan(
		&j.ID,
		&j.VideoID,
		&j.Type,
		&j.Status,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("scan job data: %w", err)
	}

	return j, nil
}
