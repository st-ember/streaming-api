package jobapp

import (
	"context"
	"fmt"
	"time"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type CompleteTranscodeJobUsecase interface {
	Execute(
		ctx context.Context,
		job *job.Job,
		result string,
		duration time.Duration,
	) error
}

type completeTranscodeJobUsecase struct {
	uowFactory repo.UnitOfWorkFactory
}

func NewCompleteTranscodeJobUsecase(uowFactory repo.UnitOfWorkFactory) *completeTranscodeJobUsecase {
	return &completeTranscodeJobUsecase{uowFactory}
}

func (u *completeTranscodeJobUsecase) Execute(
	ctx context.Context,
	job *job.Job,
	result string,
	duration time.Duration,
) error {
	// Update job entity
	if err := job.Complete(result); err != nil {
		return fmt.Errorf("complete job %s: %w", job.ID, err)
	}

	// Initialize unit of work
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Rollback(ctx)

	// Initialize repos
	videoRepo := uow.VideoRepo()
	jobRepo := uow.JobRepo()

	// Find related video
	video, err := videoRepo.FindByID(ctx, job.VideoID)
	if err != nil {
		return fmt.Errorf("get video related to job %s: %w", job.ID, err)
	}

	// Update video entity
	if err := video.UpdateDuration(duration); err != nil {
		return fmt.Errorf("update video %s duration: %w", video.ID, err)
	}
	if err := video.Publish(); err != nil {
		return fmt.Errorf("publish video %s: %w", video.ID, err)
	}

	// Persist entities
	if err := jobRepo.Save(ctx, job); err != nil {
		return fmt.Errorf("save job %s in db: %w", job.ID, err)
	}
	if err := videoRepo.Save(ctx, video); err != nil {
		return fmt.Errorf("save video %s in db: %w", video.ID, err)
	}

	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("finalize transaction %w", err)
	}

	return nil
}
