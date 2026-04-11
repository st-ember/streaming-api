package progressapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/progressstream"
	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/progress"
)

type VideoProgressUsecase interface {
	Execute(ctx context.Context, id string) (<-chan *progress.Progress, error)
}

type videoProgressUsecase struct {
	streamer   progressstream.ProgressStreamer
	uowFactory repo.UnitOfWorkFactory
}

func NewVideoProgressUsecase(
	streamer progressstream.ProgressStreamer,
	uowFactory repo.UnitOfWorkFactory,
) VideoProgressUsecase {
	return &videoProgressUsecase{
		streamer:   streamer,
		uowFactory: uowFactory,
	}
}

func (u *videoProgressUsecase) Execute(ctx context.Context, id string) (<-chan *progress.Progress, error) {
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Close(ctx)

	// Initialize repos
	jobRepo := uow.JobRepo()

	// Find related job
	j, err := jobRepo.FindByVideoID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find job with id %s: %w", id, err)
	}

	return u.streamer.Read(ctx, j.ID)
}
