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
	// Parse params
	vars := mux.Vars(r)
	resourceID := vars["resourceID"]
	filename := vars["filename"]

	// Assemble full path
	fullPath := filepath.Join(h.storagePath, resourceID, filename)

	// Ensure path validity
	cleanedPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanedPath, h.storagePath) {
		h.logger.Errorf("file path %s is invalid", fullPath)
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// Send response
	http.ServeFile(w, r, fullPath)
}
