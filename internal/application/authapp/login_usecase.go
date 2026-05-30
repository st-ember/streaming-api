package authapp

import (
	"context"
	"fmt"

	"github.com/st-ember/streaming-api/internal/application/ports/hash"
	"github.com/st-ember/streaming-api/internal/application/ports/repo"
	"github.com/st-ember/streaming-api/internal/application/ports/token"
)

type LoginUsecase interface {
	Execute(
		ctx context.Context,
		login, password string,
	) (accessToken, refreshToken string, err error)
}

type loginUsecase struct {
	authRepo repo.AuthRepo
	hasher   hash.Hasher
	token    token.Token
}

func NewLoginUsecase(
	authRepo repo.AuthRepo,
	hasher hash.Hasher,
	token token.Token,
) LoginUsecase {
	return &loginUsecase{authRepo, hasher, token}
}

func (lu *loginUsecase) Execute(
	ctx context.Context,
	key, password string,
) (accessToken, refreshToken string, err error) {
	u, err := lu.authRepo.FindUserByKey(ctx, key)
	if err != nil {
		return "", "", fmt.Errorf("find user by login %s: %w", key, err)
	}

	if !lu.hasher.Verify(password, u.PasswordHash) {
		return "", "", fmt.Errorf("invalid password: %s", password)
	}

	permissions, err := lu.authRepo.FindPermissionsByUserID(ctx, u.ID)
	if err != nil {
		return "", "", fmt.Errorf("find permissions by user id %s: %w", u.ID, err)
	}

	u.Permissions = permissions

	// Generate tokens
	at, err := lu.token.GenerateAccess(u.ID, u.Username, u.Permissions)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	rt, err := lu.token.GenerateRefresh(u.ID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return at, rt, nil
}
