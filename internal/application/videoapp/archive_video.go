package videoapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/repo"
)

// ArchiveVideoUsecase marks the entity as archived
// but does not delete the related files in storage
type ArchiveVideoUsecase interface {
	Execute(ctx context.Context, id string) error
}

type archiveVideoUsecase struct {
	uowFactory repo.UnitOfWorkFactory
}

func NewArchiveVideoUsecase(uowFactory repo.UnitOfWorkFactory) ArchiveVideoUsecase {
	return &archiveVideoUsecase{uowFactory}
}

func (u *archiveVideoUsecase) Execute(ctx context.Context, id string) error {
	uow, err := u.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return fmt.Errorf("initialize unit of work: %w", err)
	}
	defer uow.Rollback(ctx)

	videoRepo := uow.VideoRepo()

	v, err := videoRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find video %s: %w", id, err)
	}

	if err := v.Archive(); err != nil {
		return fmt.Errorf("archive video %s: %w", id, err)
	}

	if err := videoRepo.Save(ctx, v); err != nil {
		return fmt.Errorf("save video %s: %w", id, err)
	}

	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
