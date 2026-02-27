package jobapp

import (
	"context"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type FindNextPendingTranscodeJobUsecase interface {
	Execute(ctx context.Context) (*job.Job, error)
}

type findNextPendingTranscodeJobUsecase struct {
	jobRepo repo.JobRepo
}

func NewFindNextPendingTranscodeJobUsecase(jobRepo repo.JobRepo) *findNextPendingTranscodeJobUsecase {
	return &findNextPendingTranscodeJobUsecase{jobRepo}
}

func (u *findNextPendingTranscodeJobUsecase) Execute(ctx context.Context) (*job.Job, error) {
	return u.jobRepo.FindNextPendingTranscodeJob(ctx)
}
