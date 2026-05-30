package token

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	UserID      string
	Username    string
	Permissions []string
	jwt.RegisteredClaims
}

// The library will call this automatically during Parse.
func (ac AccessClaims) Validate() error {
	if ac.UserID == "" {
		return errors.New("missing user id")
	}

	return nil
}

type RefreshClaims struct {
	UserID string
	jwt.RegisteredClaims
}

// The library will call this automatically during Parse.
func (rc RefreshClaims) Validate() error {
	if rc.UserID == "" {
		return errors.New("missing user id")
	}

	return nil
}

type Token interface {
	GenerateAccess(userID, username string, permissions []string) (string, error)
	GenerateRefresh(userID string) (string, error)
	ParseAccess(token string) (*AccessClaims, error)
	ParseRefresh(token string) (*RefreshClaims, error)
}
