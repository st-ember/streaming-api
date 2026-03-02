package videoapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/domain/video"
)

type UpdateVideoUsecase interface {
	Execute(ctx context.Context, input UpdateVideoInput) (*video.Video, error)
}

type updateVideuoUsecase struct {
	uowFactory repo.UnitOfWorkFactory
}

func NewUpdateVideoUsecase(uowFactory repo.UnitOfWorkFactory) UpdateVideoUsecase {
	return &updateVideuoUsecase{uowFactory}
}

func (u *updateVideuoUsecase) Execute(ctx context.Context, input UpdateVideoInput) (*video.Video, error) {
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Rollback(ctx)

	videoRepo := uow.VideoRepo()

	v, err := videoRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("find video %s: %w", input.ID, err)
	}

	if input.Title != nil {
		if err := v.UpdateTitle(*input.Title); err != nil {
			return nil, fmt.Errorf("update video entity %s title: %w", v.ID, err)
		}
	}

	if input.Description != nil {
		if err := v.UpdateDescription(*input.Description); err != nil {
			return nil, fmt.Errorf("update video entity %s description: %w", v.ID, err)
		}
	}

	if err := videoRepo.Save(ctx, v); err != nil {
		return nil, fmt.Errorf("save video %s: %w", v.ID, err)
	}

	if err := uow.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction %s: %w", v.ID, err)
	}

	return v, nil
}
