package videoapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
)

type GetVideoInfoUsecase interface {
	Execute(ctx context.Context, id string) (*GetVideoInfoResult, error)
}

type getVideoInfoUsecase struct {
	uowFactory repo.UnitOfWorkFactory
}

func NewGetVideoInfoUsecase(uowFactory repo.UnitOfWorkFactory) GetVideoInfoUsecase {
	return &getVideoInfoUsecase{uowFactory}
}

func (u *getVideoInfoUsecase) Execute(ctx context.Context, id string) (*GetVideoInfoResult, error) {
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Rollback(ctx)

	videoRepo := uow.VideoRepo()
	jobRepo := uow.JobRepo()

	v, err := videoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find video %s: %w", id, err)
	}

	j, err := jobRepo.FindByVideoID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find job with video id %s: %w", id, err)
	}

	res := &GetVideoInfoResult{
		Video:        v,
		ManifestPath: j.Result,
		ErrorMsg:     j.ErrorMsg,
	}

	return res, nil
}
