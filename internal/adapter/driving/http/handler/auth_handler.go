package handler

import (
	"github.com/st-ember/streaming-api/internal/application/authapp"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
)

type AuthHandler struct {
	loginUC  authapp.LoginUsecase
	signupUC authapp.SignupUsecase
	logger   log.Logger
}

func NewAuthHandler(
	loginUC authapp.LoginUsecase,
	signupUC authapp.SignupUsecase,
	logger log.Logger,
) *AuthHandler {
	return &AuthHandler{loginUC, signupUC, logger}
}
