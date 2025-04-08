package handlers

import (
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

	claims := token.Claims.(jwt.MapClaims)

	// Exemple simple de réponse
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"sub": "` + claims["sub"].(string) + `"}`))
}
