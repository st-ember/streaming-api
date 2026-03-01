package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	mocklog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
	mockvideo "github.com/st-ember/streaming-api/internal/application/videoapp/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVideoHandler_Upload(t *testing.T) {
	t.Run("should return 201 Created and IDs on success", func(t *testing.T) {
		mockUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := NewVideoHandler(mockUC, mockGetInfoUC, mockLogger)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "Test Video")
		_ = writer.WriteField("description", "A test description")
		part, _ := writer.CreateFormFile("video", "test.mp4")
		_, _ = part.Write([]byte("fake-video-content"))
		_ = writer.Close()

		v, _ := video.NewVideo("vid-1", "Test Video", "Desc", "test.mp4", "res-1")
		j, _ := job.NewJob("job-1", "vid-1", job.TypeTranscode)

		mockUC.EXPECT().
			Execute(mock.Anything, mock.MatchedBy(func(in videoapp.UploadVideoInput) bool {
				return in.Title == "Test Video" && in.FileName == "test.mp4"
			})).
			Return(&videoapp.UploadVideoResult{Video: v, Job: j}, nil).
			Once()

		req := httptest.NewRequest(http.MethodPost, "/api/video/", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		h.Upload(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var resp UploadVideoResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		require.Equal(t, "vid-1", resp.VideoID)
		require.Equal(t, "job-1", resp.JobID)
	})

	t.Run("should return 400 Bad Request if multipart form is invalid", func(t *testing.T) {
		mockUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := NewVideoHandler(mockUC, mockGetInfoUC, mockLogger)

		// Send a plain text body instead of multipart
		body := bytes.NewBufferString("not a multipart form")

		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/video/", body)
		// We set the header but the body is garbage
		req.Header.Set("Content-Type", "multipart/form-data; boundary=invalid")

		w := httptest.NewRecorder()
		h.Upload(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 Bad Request if video file is missing in form", func(t *testing.T) {
		mockUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := NewVideoHandler(mockUC, mockGetInfoUC, mockLogger)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "Test Video")
		_ = writer.Close()

		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/video/", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		h.Upload(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 500 Internal Server Error if usecase fails", func(t *testing.T) {
		mockUC := mockvideo.NewMockUploadVideoUsecase(t)
		mockGetInfoUC := mockvideo.NewMockGetVideoInfoUsecase(t)
		mockLogger := mocklog.NewMockLogger(t)
		h := NewVideoHandler(mockUC, mockGetInfoUC, mockLogger)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("video", "test.mp4")
		_, _ = part.Write([]byte("content"))
		_ = writer.Close()

		mockUC.EXPECT().Execute(mock.Anything, mock.Anything).
			Return(nil, errors.New("db error")).
			Once()

		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/video/", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		h.Upload(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
