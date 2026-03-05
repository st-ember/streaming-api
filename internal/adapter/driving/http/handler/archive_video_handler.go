package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *VideoHandler) Archive(w http.ResponseWriter, r *http.Request) {
	// Parse id param
	vars := mux.Vars(r)
	id := vars["id"]

	// Execute usecase
	if err := h.archiveVideoUC.Execute(r.Context(), id); err != nil {
		h.logger.Errorf("archive video %s: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
