package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/st-ember/streaming-api/internal/adapter/driven/token"
	"github.com/st-ember/streaming-api/internal/adapter/driving/http/middleware"
	logmocks "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	tokenport "github.com/st-ember/streaming-api/internal/application/ports/token"
	tokenmocks "github.com/st-ember/streaming-api/internal/application/ports/token/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	mockToken := tokenmocks.NewMockToken(t)
	mockLogger := logmocks.NewMockLogger(t)
	mw := middleware.Auth(mockToken, mockLogger)

	// A simple final handler that verifies the context was set
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r.Context())
		if !ok || claims.UserID != "user-123" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("should succeed with valid bearer token", func(t *testing.T) {
		tokenStr := "valid-token"
		expectedClaims := &tokenport.AccessClaims{UserID: "user-123", Username: "tester"}

		mockToken.EXPECT().ParseAccess(tokenStr).Return(expectedClaims, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		rr := httptest.NewRecorder()

		mw(finalHandler).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should succeed with token in query parameter", func(t *testing.T) {
		tokenStr := "valid-query-token"
		expectedClaims := &tokenport.AccessClaims{UserID: "user-123", Username: "tester"}

		mockToken.EXPECT().ParseAccess(tokenStr).Return(expectedClaims, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/?token="+tokenStr, nil)
		rr := httptest.NewRecorder()

		mw(finalHandler).ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should fail with missing auth header", func(t *testing.T) {
		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything, "no token provided").Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		mw(finalHandler).ServeHTTP(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should fail with invalid auth format", func(t *testing.T) {
		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything, "invalid auth header format").Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic user:pass") // Wrong type
		rr := httptest.NewRecorder()

		mw(finalHandler).ServeHTTP(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should fail with invalid or expired token", func(t *testing.T) {
		tokenStr := "expired-token"
		mockToken.EXPECT().ParseAccess(tokenStr).Return(nil, token.ErrInvalidToken).Once()
		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything, "parse token: %v", mock.Anything).Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		rr := httptest.NewRecorder()

		mw(finalHandler).ServeHTTP(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
