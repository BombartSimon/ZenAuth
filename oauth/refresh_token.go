package oauth

import (
	"encoding/json"
	"net/http"
	"zenauth/oauth/store"
)

type RefreshTokenFlow struct{}

func (f *RefreshTokenFlow) Supports(grantType string) bool {
	return grantType == "refresh_token"
}

func (f *RefreshTokenFlow) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	clientID, userID, err := store.GetRefreshToken(refreshToken)
	if err != nil {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}

	var subject string
	if userID != nil {
		subject = *userID
	} else {
		subject = clientID
	}

	accessToken, err := GenerateAccessToken(subject, "default")
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
