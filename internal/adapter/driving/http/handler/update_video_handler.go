package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

func (h *VideoHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Access id param
	vars := mux.Vars(r)
	id := vars["id"]

	// Decode request
	var req UpdateVideoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf(r.Context(), "parse update request body: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Assemble input
	input := videoapp.UpdateVideoInput{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
	}

	// Execute usecase
	v, err := h.videoUC.Update.Execute(r.Context(), input)
	if err != nil {
		h.logger.Errorf(r.Context(), "update video %s: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send response
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Errorf(r.Context(), "encode video %s: %v", id, err)
	}

	// Log Success
	h.logger.Infof(r.Context(), "updated video %s", v.ID)
}
