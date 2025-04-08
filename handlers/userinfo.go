package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"zenauth/oauth"

	"github.com/golang-jwt/jwt"
)

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	token, err := oauth.ValidateAccessToken(tokenStr)
	if err != nil || !token.Valid {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["scope"] == nil {
		http.Error(w, "Scopes not available", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sub":   claims["sub"],
		"scope": claims["scope"],
	})

}
