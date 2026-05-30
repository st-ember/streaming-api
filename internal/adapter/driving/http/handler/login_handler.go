package handler

import (
	"encoding/json"
	"net/http"

	"github.com/st-ember/streaming-api/internal/application/ports/log"
)

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.logger.Errorf(r.Context(), log.CategoryAuth, "", "parse login request body: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	at, rt, err := ah.loginUC.Execute(r.Context(), req.Key, req.Password)
	if err != nil {
		ah.logger.Errorf(r.Context(), log.CategoryAuth, "", "login: %v", err)
		http.Error(w, "username, email or password was incorrect", http.StatusUnauthorized)
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
