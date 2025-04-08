package oauth

import (
	"encoding/json"
	"net/http"
	"zenauth/oauth/store"

	"github.com/google/uuid"
)

type ClientCredentialsFlow struct{}

func (f *ClientCredentialsFlow) Supports(grantType string) bool {
	return grantType == "client_credentials"
}

func (f *ClientCredentialsFlow) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "invalid_client", http.StatusUnauthorized)
		return
	}

	client, err := store.GetClientByID(clientID)
	if err != nil || client.Secret != clientSecret {
		http.Error(w, "invalid_client", http.StatusUnauthorized)
		return
	}

	accessToken, err := GenerateAccessToken(clientID, "default")
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}

	refreshToken := uuid.NewString()
	_ = store.StoreRefreshToken(refreshToken, clientID, nil)

	token := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}
