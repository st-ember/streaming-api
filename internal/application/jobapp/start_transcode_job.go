package jobapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type StartTranscodeJobUsecase interface {
	Execute(ctx context.Context, job *job.Job) (*StartTranscodeJobResponse, error)
}

type startTranscodeJobUsecase struct {
	uowFactory repo.UnitOfWorkFactory
}

func NewStartTranscodeJobUsecase(
	uowFactory repo.UnitOfWorkFactory,
) *startTranscodeJobUsecase {
	return &startTranscodeJobUsecase{
		uowFactory,
	}
}

func (u *startTranscodeJobUsecase) Execute(ctx context.Context, job *job.Job) (*StartTranscodeJobResponse, error) {
	// Update job entity
	if err := job.Start(); err != nil {
		return nil, fmt.Errorf("start job %s: %w", job.ID, err)
	}

	// Initialize unit of work
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Rollback(ctx)

	// Initialize repos
	videoRepo := uow.VideoRepo()
	jobRepo := uow.JobRepo()

	// Find related video
	video, err := videoRepo.FindByID(ctx, job.VideoID)
	if err != nil {
		return nil, fmt.Errorf("get video related to job %s: %w", job.ID, err)
	}

	// Update video entity
	if err := video.MarkAsProcessing(); err != nil {
		return nil, fmt.Errorf("mark video %s as processing: %w", video.ID, err)
	}

	// Persist entities
	if err := jobRepo.Save(ctx, job); err != nil {
		return nil, fmt.Errorf("save job %s in db: %w", job.ID, err)
	}
	if err := videoRepo.Save(ctx, video); err != nil {
		return nil, fmt.Errorf("save video %s in db: %w", video.ID, err)
	}

	if err := uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("finalize transaction %w", err)
	}

	return &StartTranscodeJobResponse{
		ResourceID:     video.ResourceID,
		SourceFilename: video.Filename,
	}, nil
}
