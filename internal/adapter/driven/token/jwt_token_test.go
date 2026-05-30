package token_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/st-ember/streaming-api/internal/adapter/driven/token"
	"github.com/stretchr/testify/require"
)

type jwtTokenTestHelper struct {
	accessSecret  []byte
	refreshSecret []byte
	require       *require.Assertions
}

func setupJwtTokenTestHelper(t *testing.T) *jwtTokenTestHelper {
	return &jwtTokenTestHelper{
		accessSecret:  []byte("access-secret"),
		refreshSecret: []byte("refresh-secret"),
		require:       require.New(t),
	}
}

func TestAccessSuite(t *testing.T) {
	h := setupJwtTokenTestHelper(t)
	jwtToken := token.NewJwtToken(h.accessSecret, h.refreshSecret)

	t.Run("should generate and parse access token", func(t *testing.T) {
		userID := "user-123"
		username := "tester"
		permissions := []string{"video:upload", "video:delete"}

		tokenStr, err := jwtToken.GenerateAccess(userID, username, permissions)
		h.require.NoError(err)
		h.require.NotEmpty(tokenStr)

		claims, err := jwtToken.ParseAccess(tokenStr)
		h.require.NoError(err)
		h.require.Equal(userID, claims.UserID)
		h.require.Equal(username, claims.Username)
		h.require.ElementsMatch(permissions, claims.Permissions)
	})

	t.Run("should fail to parse access token with wrong secret", func(t *testing.T) {
		tgWrong := token.NewJwtToken([]byte("wrong-secret"), h.refreshSecret)

		tokenStr, _ := jwtToken.GenerateAccess("id", "user", nil)
		_, err := tgWrong.ParseAccess(tokenStr)
		h.require.ErrorIs(err, token.ErrInvalidToken)
	})

	t.Run("should fail with invalid token string", func(t *testing.T) {
		_, err := jwtToken.ParseAccess("invalid-token")
		h.require.ErrorIs(err, token.ErrInvalidToken)
	})

	t.Run("should fail with mismatched payload", func(t *testing.T) {
		wrongToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"wrong_field": "data",
		})
		signedStr, _ := wrongToken.SignedString(h.accessSecret)

		_, err := jwtToken.ParseAccess(signedStr)

		h.require.ErrorIs(err, token.ErrInvalidToken)
		h.require.ErrorContains(err, "missing user id")
	})

	t.Run("should fail with invalid signing method", func(t *testing.T) {
		invalidAlgToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyJ9.fake-sig"
		_, err := jwtToken.ParseAccess(invalidAlgToken)

		h.require.ErrorIs(err, token.ErrInvalidSigningMethod)
	})
}

func TestRefreshSuite(t *testing.T) {
	h := setupJwtTokenTestHelper(t)
	jwtToken := token.NewJwtToken(h.accessSecret, h.refreshSecret)

	t.Run("should generate and parse refresh token", func(t *testing.T) {
		userID := "user-123"

		tokenStr, err := jwtToken.GenerateRefresh(userID)
		h.require.NoError(err)
		h.require.NotEmpty(tokenStr)

		claims, err := jwtToken.ParseRefresh(tokenStr)
		h.require.NoError(err)
		h.require.Equal(userID, claims.UserID)
	})

	t.Run("should fail to parse refresh token with wrong secret", func(t *testing.T) {
		tgWrong := token.NewJwtToken(h.accessSecret, []byte("wrong-secret"))

		tokenStr, _ := jwtToken.GenerateRefresh("id")
		_, err := tgWrong.ParseRefresh(tokenStr)
		h.require.ErrorIs(err, token.ErrInvalidToken)
	})

	t.Run("should fail with invalid token string", func(t *testing.T) {
		_, err := jwtToken.ParseRefresh("invalid-token")
		h.require.ErrorIs(err, token.ErrInvalidToken)
	})

	t.Run("should fail with mismatched payload", func(t *testing.T) {
		wrongToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"wrong_field": "data",
		})
		signedStr, _ := wrongToken.SignedString(h.refreshSecret)

		_, err := jwtToken.ParseRefresh(signedStr)

		h.require.ErrorIs(err, token.ErrInvalidToken)
		h.require.ErrorContains(err, "missing user id")
	})

	t.Run("should fail with invalid signing method", func(t *testing.T) {
		invalidAlgToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyJ9.fake-sig"
		_, err := jwtToken.ParseRefresh(invalidAlgToken)

		h.require.ErrorIs(err, token.ErrInvalidSigningMethod)
	})
}
