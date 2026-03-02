package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	mocklog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
	mockvideo "github.com/st-ember/streaming-api/internal/application/videoapp/mocks"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVideoHandler_Update(t *testing.T) {
	t.Run("should return 200 OK and video info on success", func(t *testing.T) {
		mockUploadUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockUpdateUC := mockvideo.NewMockUpdateVideoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := NewVideoHandler(mockUploadUC, mockGetInfoUC, mockUpdateUC, mockLogger)

		// Video that will be returned by usecase
		videoID := "video-123"
		resourceID := "resource-123"

		newTitle := "new_title"
		updatedVideo, err := video.NewVideo(videoID, newTitle, "Description", "test.mp4", resourceID)
		require.NoError(t, err)

		updateInput := videoapp.UpdateVideoInput{
			ID:    videoID,
			Title: &newTitle,
		}

		mockUpdateUC.EXPECT().
			Execute(mock.Anything, updateInput).
			Return(updatedVideo, nil).
			Once()

		// Prepare request body
		updateReq := UpdateVideoRequest{
			Title: &newTitle,
		}
		body, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/api/video/"+videoID, bytes.NewReader(body))

		// Manually set gorilla/mux vars
		req = mux.SetURLVars(req, map[string]string{"id": videoID})

		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp *video.Video
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, videoID, resp.ID)
		require.Equal(t, newTitle, resp.Title)
	})

	t.Run("should return 500 Internal Server Error if usecase fails", func(t *testing.T) {
		mockUploadUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockUpdateUC := mockvideo.NewMockUpdateVideoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := NewVideoHandler(mockUploadUC, mockGetInfoUC, mockUpdateUC, mockLogger)

		videoID := "video-123"
		updateInput := videoapp.UpdateVideoInput{
			ID: videoID,
		}
		mockUpdateUC.EXPECT().
			Execute(mock.Anything, updateInput).
			Return(nil, errors.New("db failure")).
			Once()

		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		// Prepare request body
		updateReq := UpdateVideoRequest{}
		body, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/api/video/"+videoID, bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"id": videoID})
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
