package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *VideoHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse page param
	vars := mux.Vars(r)
	pageStr := vars["page"]
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Errorf(r.Context(), "parse page param %s: %v", pageStr, err)
		http.Error(w, "invalid page param", http.StatusBadRequest)
		return
	}

	// Execute usecase
	vs, err := h.videoUC.List.Execute(r.Context(), page)
	if err != nil {
		h.logger.Errorf(r.Context(), "list videos: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(vs); err != nil {
		h.logger.Errorf(r.Context(), "encode video list: %v", err)
	}

	// Log success
	h.logger.Infof(r.Context(), "listed videos")
}
