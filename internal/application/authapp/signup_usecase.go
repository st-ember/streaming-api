package authapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/st-ember/streaming-api/internal/application/ports/hash"
	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/application/ports/token"
	"github.com/st-ember/streaming-api/internal/domain/user"
)

type SignupUsecase interface {
	Execute(
		ctx context.Context,
		username, email, password, roleName string,
	) (accessToken, refreshToken string, err error)
}

type signupUsecase struct {
	uowFactory repo.UnitOfWorkFactory
	authRepo   repo.AuthRepo
	hasher     hash.Hasher
	token      token.Token
}

func NewSignupUsecase(
	uowFactory repo.UnitOfWorkFactory,
	authRepo repo.AuthRepo,
	hasher hash.Hasher,
	token token.Token,
) SignupUsecase {
	return &signupUsecase{uowFactory, authRepo, hasher, token}
}

func (su *signupUsecase) Execute(
	ctx context.Context,
	username, email, password, roleName string,
) (accessToken, refreshToken string, err error) {
	hashedPwd, err := su.hasher.Hash(password)
	if err != nil {
		return "", "", fmt.Errorf("hash password")
	}

	userID := uuid.NewString()
	u, err := user.NewUser(userID, email, username, hashedPwd)
	if err != nil {
		return "", "", fmt.Errorf("create user: %w", err)
	}

	permissions, err := su.authRepo.FindPermissionsByRole(ctx, roleName)
	if err != nil {
		return "", "", fmt.Errorf("find permissions for rolename %s: %w", roleName, err)
	}

	// Assign permissions
	u.Permissions = permissions

	// Acquire auth repo with transaction
	uow, err := su.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return "", "", fmt.Errorf("initialize unit of work")
	}
	defer uow.Close(ctx)

	txAuthRepo := uow.AuthRepo()

	if err := txAuthRepo.SaveUser(ctx, u); err != nil {
		uow.Rollback(ctx)
		return "", "", fmt.Errorf("save user: %w", err)
	}

	if err := txAuthRepo.SaveUserRole(ctx, userID, roleName); err != nil {
		uow.Rollback(ctx)
		return "", "", fmt.Errorf("save user role: %w", err)
	}

	if err := uow.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("commit user info: %w", err)
	}

	// Generate tokens
	at, err := su.token.GenerateAccess(u.ID, u.Username, u.Permissions)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	rt, err := su.token.GenerateRefresh(u.ID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return at, rt, nil
}
