package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
)

func (h *VideoHandler) Archive(w http.ResponseWriter, r *http.Request) {
	// Parse id param
	vars := mux.Vars(r)
	id := vars["id"]

	// Execute usecase
	if err := h.videoUC.Archive.Execute(r.Context(), id); err != nil {
		h.logger.Errorf(r.Context(), log.CategoryVideo, id, "archive video %s: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Send OK response
	w.WriteHeader(http.StatusOK)

	// Log success
	h.logger.Infof(r.Context(), log.CategoryVideo, id, "archived video %s", id)
}
