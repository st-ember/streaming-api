package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *VideoHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse page param
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Errorf("parse page param %s: %v", pageStr, err)
		http.Error(w, "invalid page param", http.StatusBadRequest)
		return
	}

	vs, err := h.videoUC.List.Execute(r.Context(), page)
	if err != nil {
		h.logger.Errorf("list videos: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(vs); err != nil {
		h.logger.Errorf("encode video list: %v", err)
	}
}
