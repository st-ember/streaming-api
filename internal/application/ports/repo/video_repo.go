package repo

import (
	"context"

	"github.com/st-ember/streaming-api/internal/domain/video"
)

type VideoRepo interface {
	Save(ctx context.Context, video *video.Video) error
	FindByID(ctx context.Context, id string) (*video.Video, error)
}
