package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
)

func (h *VideoHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Parse id param
	vars := mux.Vars(r)
	id := vars["id"]

	// Execute usecase
	info, err := h.videoUC.GetInfo.Execute(r.Context(), id)
	if err != nil {
		h.logger.Errorf(r.Context(), log.CategoryVideo, id, "find video %s: %v", id, err)
		http.Error(w, "failed to find video info", http.StatusInternalServerError)
		return
	}

	// Assemble response
	res := GetVideoInfoResponse{
		ID:             info.Video.ID,
		Title:          info.Video.Title,
		Description:    info.Video.Description,
		SourceFilename: info.Video.Filename,
		ResourceID:     info.Video.ResourceID,
		Status:         string(info.Video.Status),
		Duration:       info.Video.Duration.Seconds(),
		ManifestPath:   info.ManifestPath,
		ErrorMsg:       info.ErrorMsg,
		CreatedAt:      info.Video.CreatedAt,
		UpdatedAt:      info.Video.UpdatedAt,
	}

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send response
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Errorf(r.Context(), log.CategoryVideo, info.Video.ID, "encode get video info response %v", err)
	}

	// Log success
	h.logger.Infof(r.Context(), log.CategoryVideo, info.Video.ID, "got video %s info", id)
}
