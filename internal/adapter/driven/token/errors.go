package token

import "errors"

var (
	ErrInvalidSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken         = errors.New("invalid token")
)
