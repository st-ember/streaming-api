package videoapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/application/ports/storage"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
)

type UploadVideoUsecase struct {
	assetStorer storage.AssetStorer
	uowFactory  repo.UnitOfWorkFactory
	logger      log.Logger
}

func NewUploadVideoUsecase(
	assetStorer storage.AssetStorer,
	uow repo.UnitOfWorkFactory,
	logger log.Logger,
) *UploadVideoUsecase {
	return &UploadVideoUsecase{
		assetStorer,
		uow,
		logger,
	}
}

func (u *UploadVideoUsecase) Execute(ctx context.Context, input UploadVideoInput) (*UploadVideoResponse, error) {
	// store original video
	resourceID := uuid.NewString()
	err := u.assetStorer.Save(ctx, resourceID, input.FileName, input.VideoContent)
	if err != nil {
		return nil, fmt.Errorf("store asset %s: %w", resourceID, err)
	}

	// defer cleanup on error
	defer func() {
		if err != nil {
			if cleanupErr := u.assetStorer.DeleteAll(ctx, resourceID); cleanupErr != nil {
				u.logger.Errorf("Failed to clean up: %v", cleanupErr)
			}
		}
	}()

	// create video entity
	videoID := uuid.NewString()
	v, err := video.NewVideo(videoID, input.Title, input.Description, input.FileName, resourceID)
	if err != nil {
		return nil, fmt.Errorf("create new video %s: %w", videoID, err)
	}

	// create job entity
	jobID := uuid.NewString()
	j, err := job.NewJob(jobID, videoID, job.TypeTranscode)
	if err != nil {
		return nil, fmt.Errorf("create new job %s: %w", jobID, err)
	}

	// initialize unit of work
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Rollback(ctx)

	// initialize repos
	videoRepo := uow.VideoRepo()
	jobRepo := uow.JobRepo()

	// save to video repo
	err = videoRepo.Save(ctx, v)
	if err != nil {
		return nil, fmt.Errorf("save video %s in db: %w", videoID, err)
	}

	// save to job repo
	err = jobRepo.Save(ctx, j)
	if err != nil {
		return nil, fmt.Errorf("save job %s in db: %w", jobID, err)
	}

	err = uow.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("finalize transaction: %w", err)
	}

	return &UploadVideoResponse{Video: v, Job: j}, nil
}
