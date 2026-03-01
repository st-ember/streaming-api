package jobapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type FindNextPendingTranscodeJobUsecase interface {
	Execute(ctx context.Context) (*job.Job, error)
}

type findNextPendingTranscodeJobUsecase struct {
	uowFactory repo.UnitOfWorkFactory
}

func NewFindNextPendingTranscodeJobUsecase(uowFactory repo.UnitOfWorkFactory) *findNextPendingTranscodeJobUsecase {
	return &findNextPendingTranscodeJobUsecase{uowFactory}
}

func (u *findNextPendingTranscodeJobUsecase) Execute(ctx context.Context) (*job.Job, error) {
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Rollback(ctx)

	jobRepo := uow.JobRepo()
	return jobRepo.FindNextPendingTranscodeJob(ctx)
}
