package videoapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/video"
)

type ListVideosUsecase interface {
	Execute(ctx context.Context, page int) ([]*video.Video, error)
}

type listVideoUsecase struct {
	uowFactory repo.UnitOfWorkFactory
}

func NewListVideoUsecase(uowFactory repo.UnitOfWorkFactory) ListVideosUsecase {
	return &listVideoUsecase{uowFactory}
}

func (u *listVideoUsecase) Execute(ctx context.Context, page int) ([]*video.Video, error) {
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize unit of work")
	}
	defer uow.Rollback(ctx)

	videoRepo := uow.VideoRepo()

	return videoRepo.List(ctx, page)
}
