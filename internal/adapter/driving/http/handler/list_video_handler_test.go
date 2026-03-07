package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/st-ember/streaming-api/internal/adapter/driving/http/handler"
	mocklog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
	mockvideo "github.com/st-ember/streaming-api/internal/application/videoapp/mocks"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVideoHandler_List(t *testing.T) {
	t.Run("should return 200 OK and video list on success", func(t *testing.T) {
		mockListUC := mockvideo.NewMockListVideosUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)

		videoUCs := videoapp.VideoUsecase{List: mockListUC}
		h := handler.NewVideoHandler(videoUCs, mockLogger)

		page := 1
		expectedVideos := []*video.Video{
			{ID: "video-1", Title: "First"},
			{ID: "video-2", Title: "Second"},
		}

		mockListUC.EXPECT().Execute(mock.Anything, page).Return(expectedVideos, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/video/?page=1", nil)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var resp []*video.Video
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp, 2)
	})

	t.Run("should return 400 Bad Request on invalid page param", func(t *testing.T) {
		mockLogger := mocklog.NewMockLogger(t)
		h := handler.NewVideoHandler(videoapp.VideoUsecase{}, mockLogger)

		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/video/?page=abc", nil)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should return 500 Internal Server Error if usecase fails", func(t *testing.T) {
		mockListUC := mockvideo.NewMockListVideosUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)

		videoUCs := videoapp.VideoUsecase{List: mockListUC}
		h := handler.NewVideoHandler(videoUCs, mockLogger)

		mockListUC.EXPECT().Execute(mock.Anything, 1).Return(nil, errors.New("db fail")).Once()
		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/video/?page=1", nil)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
