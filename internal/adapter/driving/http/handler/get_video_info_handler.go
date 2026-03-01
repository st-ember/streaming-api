package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (h *VideoHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Parse id param
	vars := mux.Vars(r)
	id := vars["id"]

	// Execute usecase
	info, err := h.getVideoUC.Execute(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to find video info", http.StatusInternalServerError)
		h.logger.Errorf("find video %s: %v", err)
		return
	}

	// Assemble response
	res := GetVideoInfoResponse{
		ID:             info.Video.ID,
		Title:          info.Video.Title,
		Description:    info.Video.Description,
		SourceFilename: info.Video.Filename,
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
		h.logger.Errorf("encode get video info response %v", err)
	}
}
