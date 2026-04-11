package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/progressapp"
	"github.com/st-ember/streaming-api/internal/domain/progress"
)

type ProgressHandler struct {
	videoProgressUC progressapp.VideoProgressUsecase
	logger          log.Logger
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewProgressHandler(
	videoProgressUC progressapp.VideoProgressUsecase,
	logger log.Logger,
) *ProgressHandler {
	return &ProgressHandler{
		videoProgressUC: videoProgressUC,
		logger:          logger,
	}
}

func (h *ProgressHandler) VideoProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	prgCh, err := h.videoProgressUC.Execute(r.Context(), id)
	if err != nil {
		h.logger.Errorf(r.Context(), "execute video progress usecase: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "video not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf(r.Context(), "upgrade to websocket connection: %v", err)
		return
	}
	defer conn.Close()

	for {
		select {
		case <-r.Context().Done():
			return
		case prg, ok := <-prgCh:
			if !ok { // channel closed
				return
			}

			if err := conn.WriteJSON(prg); err != nil {
				h.logger.Errorf(r.Context(), "write json to websocket connection: %v", err)
				return
			}

			if prg.Status == progress.StatusEnd || prg.Status == progress.StatusError {
				return
			}
		}
	}
}
