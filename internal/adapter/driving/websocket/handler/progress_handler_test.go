package handler_test

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/st-ember/streaming-api/internal/adapter/driving/websocket/handler"
	logmocks "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	progressappmocks "github.com/st-ember/streaming-api/internal/application/progressapp/mocks"
	"github.com/st-ember/streaming-api/internal/domain/progress"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProgressHandler(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		t.Parallel()

		// --- ARRANGE ---
		mockUC := progressappmocks.NewMockVideoProgressUsecase(t)
		mockLogger := logmocks.NewMockLogger(t)
		progressHandler := handler.NewProgressHandler(mockUC, mockLogger)

		videoID := "test-video-id"

		// Setup progress channel and updates
		prgCh := make(chan *progress.Progress, 2)

		p1, _ := progress.NewProgress(100)
		_ = p1.UpdateCurrentFrames(50)

		p2, _ := progress.NewProgress(100)
		_ = p2.UpdateCurrentFrames(100)
		_ = p2.End()

		prgCh <- p1
		prgCh <- p2
		close(prgCh)

		mockUC.EXPECT().Execute(mock.Anything, videoID).Return(prgCh, nil).Once()

		// Set up mux for gorilla/mux to parse {id}
		r := mux.NewRouter()
		r.HandleFunc("/progress/video/{id}", progressHandler.VideoProgress)

		server := httptest.NewServer(r)
		defer server.Close()

		// Convert http URL to ws URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/progress/video/" + videoID
		// --- ACT ---
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// --- ASSERT ---
		var receivedP1 progress.Progress
		err = conn.ReadJSON(&receivedP1)
		require.NoError(t, err)
		require.Equal(t, int64(50), receivedP1.CurrentFrames)

		var receivedP2 progress.Progress
		err = conn.ReadJSON(&receivedP2)
		require.NoError(t, err)
		require.Equal(t, progress.StatusEnd, receivedP2.Status)
	})

	t.Run("video not found", func(t *testing.T) {
		t.Parallel()

		// --- ARRANGE ---
		mockUC := progressappmocks.NewMockVideoProgressUsecase(t)
		mockLogger := logmocks.NewMockLogger(t)
		progressHandler := handler.NewProgressHandler(mockUC, mockLogger)

		videoID := "non-existent-id"
		mockUC.EXPECT().Execute(mock.Anything, videoID).Return(nil, sql.ErrNoRows).Once()
		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()

		r := mux.NewRouter()
		r.HandleFunc("/videos/{id}/progress", progressHandler.VideoProgress)

		server := httptest.NewServer(r)
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/videos/" + videoID + "/progress"

		// --- ACT ---
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// --- ASSERT ---
		require.Error(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("internal server error", func(t *testing.T) {
		t.Parallel()

		// --- ARRANGE ---
		mockUC := progressappmocks.NewMockVideoProgressUsecase(t)
		mockLogger := logmocks.NewMockLogger(t)
		progressHandler := handler.NewProgressHandler(mockUC, mockLogger)

		videoID := "error-id"
		mockUC.EXPECT().Execute(mock.Anything, videoID).Return(nil, errors.New("db error")).Once()
		mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()

		r := mux.NewRouter()
		r.HandleFunc("/videos/{id}/progress", progressHandler.VideoProgress)

		server := httptest.NewServer(r)
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/videos/" + videoID + "/progress"

		// --- ACT ---
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// --- ASSERT ---
		require.Error(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("terminates on progress error status", func(t *testing.T) {
		t.Parallel()

		// --- ARRANGE ---
		mockUC := progressappmocks.NewMockVideoProgressUsecase(t)
		mockLogger := logmocks.NewMockLogger(t)
		progressHandler := handler.NewProgressHandler(mockUC, mockLogger)

		videoID := "video-id"
		prgCh := make(chan *progress.Progress, 1)
		p, _ := progress.NewProgress(100)
		_ = p.MarkAsError()
		prgCh <- p
		close(prgCh)

		mockUC.EXPECT().Execute(mock.Anything, videoID).Return(prgCh, nil).Once()

		r := mux.NewRouter()
		r.HandleFunc("/videos/{id}/progress", progressHandler.VideoProgress)

		server := httptest.NewServer(r)
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/videos/" + videoID + "/progress"

		// --- ACT ---
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// --- ASSERT ---
		var receivedP progress.Progress
		err = conn.ReadJSON(&receivedP)
		require.NoError(t, err)
		require.Equal(t, progress.StatusError, receivedP.Status)

		// Verification that connection closes: the next read should return an error
		_, _, err = conn.ReadMessage()
		require.Error(t, err)
	})
}
