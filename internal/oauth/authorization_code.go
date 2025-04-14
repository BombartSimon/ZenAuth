package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"zenauth/internal/repositories"

	"github.com/google/uuid"
)

type AuthorizationCodeFlow struct{}

func (f *AuthorizationCodeFlow) Supports(grantType string) bool {
	return grantType == "authorization_code"
}

func (f *AuthorizationCodeFlow) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" || redirectURI == "" || codeVerifier == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	// Retrieve the code from the database
	authCode, err := repositories.GetAuthCode(code)
	if err != nil || time.Now().After(authCode.ExpiresAt) {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}

	client, err := repositories.GetClientByID(authCode.ClientID)
	if err != nil {
		http.Error(w, "unauthorized_client", http.StatusBadRequest)
		return
	}

	if !isRedirectURIAuthorized(redirectURI, client.RedirectURIs) {
		http.Error(w, "invalid_redirect_uri", http.StatusBadRequest)
		return
	}

	// Verify redirect_uri
	if authCode.RedirectURI != redirectURI {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}

	// Verify PKCE
	if err := verifyPKCE(authCode.CodeChallenge, authCode.CodeChallengeMethod, codeVerifier); err != nil {
		http.Error(w, "invalid_grant (pkce)", http.StatusBadRequest)
		return
	}

	// Delete the code after use (security)
	_ = repositories.DeleteAuthCode(code)

	// Generate access_token
	accessToken, err := GenerateAccessToken(authCode.UserID, authCode.Scope)
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}

	// Generate refresh_token
	refreshToken := generateRandomToken()
	_ = repositories.StoreRefreshToken(refreshToken, authCode.ClientID, &authCode.UserID)

	// Response
	token := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func verifyPKCE(challenge, method, verifier string) error {
	if challenge == "" || method == "" {
		return nil
	}

	switch method {
	case "S256":
		h := sha256.Sum256([]byte(verifier))
		encoded := base64.RawURLEncoding.EncodeToString(h[:])
		if encoded != challenge {
			return errors.New("PKCE S256 verification failed")
		}
	case "plain":
		if verifier != challenge {
			return errors.New("PKCE plain verification failed")
		}
	default:
		return errors.New("unsupported PKCE method")
	}
	return nil
}

func generateRandomToken() string {
	return uuid.NewString()
}

func isRedirectURIAuthorized(uri string, allowed []string) bool {
	for _, u := range allowed {
		if u == uri {
			return true
		}
	}
	return false
}
