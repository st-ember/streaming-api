package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
)

type StreamingHandler struct {
	storagePath string
	logger      log.Logger
}

func NewStreamingHandler(storagePath string, logger log.Logger) *StreamingHandler {
	return &StreamingHandler{storagePath, logger}
}

func (h *StreamingHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceID := vars["resourceID"]
	filename := vars["filename"]

	fullPath := filepath.Join(h.storagePath, resourceID, filename)
	cleanedPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanedPath, h.storagePath) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, fullPath)
}
