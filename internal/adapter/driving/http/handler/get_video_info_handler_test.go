package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/adapter/driving/http/handler"
	mocklog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
	mockvideo "github.com/st-ember/streaming-api/internal/application/videoapp/mocks"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVideoHandler_Get(t *testing.T) {
	t.Run("should return 200 OK and video info on success", func(t *testing.T) {
		mockUploadUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := handler.NewVideoHandler(mockUploadUC, mockGetInfoUC, mockLogger)

		videoID := "video-123"
		resourceID := "resource-123"
		v, _ := video.NewVideo(videoID, "Test Video", "Description", "test.mp4", resourceID)
		v.Duration = 120 * time.Second

		usecaseResult := &videoapp.GetVideoInfoResult{
			Video:        v,
			ManifestPath: "storage/video-123/manifest.m3u8",
			ErrorMsg:     "",
		}

		mockGetInfoUC.EXPECT().
			Execute(mock.Anything, videoID).
			Return(usecaseResult, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/api/video/"+videoID, nil)
		// Manually set gorilla/mux vars
		req = mux.SetURLVars(req, map[string]string{"id": videoID})

		rr := httptest.NewRecorder()

		h.Get(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp handler.GetVideoInfoResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, videoID, resp.ID)
		require.Equal(t, "Test Video", resp.Title)
		require.Equal(t, 120.0, resp.Duration)
		require.Equal(t, usecaseResult.ManifestPath, resp.ManifestPath)
	})

	t.Run("should return 500 Internal Server Error if usecase fails", func(t *testing.T) {
		mockUploadUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := handler.NewVideoHandler(mockUploadUC, mockGetInfoUC, mockLogger)

		videoID := "video-123"
		mockGetInfoUC.EXPECT().
			Execute(mock.Anything, videoID).
			Return(nil, errors.New("db failure")).
			Once()

		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/video/"+videoID, nil)
		req = mux.SetURLVars(req, map[string]string{"id": videoID})
		rr := httptest.NewRecorder()

		h.Get(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
