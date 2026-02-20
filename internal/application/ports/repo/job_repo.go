package repo

import (
	"context"

	"github.com/st-ember/streaming-api/internal/domain/job"
)

type JobRepo interface {
	Save(ctx context.Context, job *job.Job) error
}
