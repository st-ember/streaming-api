package progressstream

import (
	"context"

	"github.com/st-ember/streaming-api/internal/domain/progress"
)

type ProgressStreamer interface {
	Push(ctx context.Context, jobID string, prg *progress.Progress) error
	Read(ctx context.Context, jobID string) (<-chan *progress.Progress, error)
}
