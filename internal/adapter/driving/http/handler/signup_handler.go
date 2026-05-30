package handler

import (
	"encoding/json"
	"net/http"

	"github.com/st-ember/streaming-api/internal/application/ports/log"
)

func (ah *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.logger.Errorf(r.Context(), log.CategoryAuth, "", "parse signup request body: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	at, rt, err := ah.signupUC.Execute(r.Context(), req.Username, req.Email, req.Password, req.RoleName)
	if err != nil {
		ah.logger.Errorf(r.Context(), log.CategoryAuth, "", "signup: %v", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    rt,
		HttpOnly: true,
		Path:     "/refresh",
		Secure:   false, // temp for testing
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": at,
	})
}
