package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/st-ember/streaming-api/internal/application/ports/token"
)

type JwtToken struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJwtToken(accessSecret, refreshSecret []byte) token.Token {
	return &JwtToken{accessSecret, refreshSecret}
}

func (tg *JwtToken) GenerateAccess(userID, username string, permissions []string) (string, error) {
	claims := token.AccessClaims{
		UserID:      userID,
		Username:    username,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tg.accessSecret)
}

func (tg *JwtToken) GenerateRefresh(userID string) (string, error) {
	claims := token.RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tg.refreshSecret)
}

func (tg *JwtToken) ParseAccess(tokenStr string) (*token.AccessClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &token.AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, t.Header["alg"])
		}

		return tg.accessSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if !t.Valid {
		return nil, ErrInvalidToken
	}

	// Will not fail as type has been declared in ParseWithClaims
	claims, _ := t.Claims.(*token.AccessClaims)

	return claims, nil
}

func (tg *JwtToken) ParseRefresh(tokenStr string) (*token.RefreshClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &token.RefreshClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, t.Header["alg"])
		}

		return tg.refreshSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if !t.Valid {
		return nil, ErrInvalidToken
	}

	// Will not fail as type has been declared in ParseWithClaims
	claims, _ := t.Claims.(*token.RefreshClaims)

	return claims, nil
}
