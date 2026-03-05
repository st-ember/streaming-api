package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/adapter/driving/http/handler"
	mocklog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	mockvideo "github.com/st-ember/streaming-api/internal/application/videoapp/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVideoHandler_Archive(t *testing.T) {
	t.Run("should return 200 OK on success", func(t *testing.T) {
		mockUploadUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockUpdateUC := mockvideo.NewMockUpdateVideoUsecase(t)
		mockArchiveUC := mockvideo.NewMockArchiveVideoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := handler.NewVideoHandler(mockUploadUC, mockGetInfoUC, mockUpdateUC, mockArchiveUC, mockLogger)

		videoID := "video-123"

		mockArchiveUC.EXPECT().
			Execute(mock.Anything, videoID).
			Return(nil).
			Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/video/"+videoID, nil)
		// Manually set gorilla/mux vars
		req = mux.SetURLVars(req, map[string]string{"id": videoID})

		rr := httptest.NewRecorder()

		h.Archive(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 500 Internal Server Error if usecase fails", func(t *testing.T) {
		mockUploadUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockUpdateUC := mockvideo.NewMockUpdateVideoUsecase(t)
		mockArchiveUC := mockvideo.NewMockArchiveVideoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := handler.NewVideoHandler(mockUploadUC, mockGetInfoUC, mockUpdateUC, mockArchiveUC, mockLogger)

		videoID := "video-123"
		mockArchiveUC.EXPECT().
			Execute(mock.Anything, videoID).
			Return(errors.New("db failure")).
			Once()

		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/video/"+videoID, nil)
		req = mux.SetURLVars(req, map[string]string{"id": videoID})
		rr := httptest.NewRecorder()

		h.Archive(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
