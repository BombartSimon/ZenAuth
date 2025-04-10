package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"zenauth/config"
	adapters "zenauth/internal/adapters/auth_providers"
	userAdapters "zenauth/internal/adapters/users"
	"zenauth/internal/models"
	"zenauth/internal/repositories"

	"github.com/google/uuid"
)

// StartExternalAuth initiates OAuth flow with the specified provider
func StartExternalAuth(w http.ResponseWriter, r *http.Request) {
	// Get provider ID from query parameters
	providerID := r.URL.Query().Get("provider")
	if providerID == "" {
		http.Error(w, "provider parameter is required", http.StatusBadRequest)
		return
	}

	// Validate the redirect URL for the OAuth flow
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		http.Error(w, "redirect_uri parameter is required", http.StatusBadRequest)
		return
	}

	// Get client ID
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		http.Error(w, "client_id parameter is required", http.StatusBadRequest)
		return
	}

	// Generate state parameter to prevent CSRF
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(b)

	// Store the state with original OAuth parameters
	// This would normally go in a session or temporary storage
	// For simplicity, we're storing it in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "client_id",
		Value:    clientID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "redirect_uri",
		Value:    redirectURI,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	// Get the provider configuration
	provider, err := repositories.GetAuthProviderByID(providerID)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	if !provider.Enabled {
		http.Error(w, "This authentication provider is disabled", http.StatusBadRequest)
		return
	}

	// Get the OAuth2 provider implementation
	oauthProvider, err := adapters.GetProvider(provider)
	if err != nil {
		http.Error(w, "Failed to initialize provider", http.StatusInternalServerError)
		return
	}

	var authURL string
	if provider.Type == models.MicrosoftProvider {
		tenantID := provider.TenantID
		if tenantID == "" {
			tenantID = "organizations"
		}

		u, _ := url.Parse(fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", tenantID))
		q := u.Query()
		q.Set("client_id", provider.ClientID)
		q.Set("response_type", "code")
		q.Set("redirect_uri", fmt.Sprintf("http://%s/auth/callback/%s", r.Host, providerID))
		q.Set("state", state)
		q.Set("scope", "openid email profile User.Read")

		u.RawQuery = q.Encode()
		authURL = u.String()

	} else {
		// For other providers, use the generic OAuth2 implementation
		authURL = oauthProvider.GetAuthURL(state, fmt.Sprintf("http://%s/auth/callback/%s", r.Host, providerID))
	}

	log.Printf("Microsoft Auth URL: %s", authURL)
	log.Printf("Provider details: Type=%s, TenantID=%s", provider.Type, provider.TenantID)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleExternalAuthCallback processes OAuth callback from providers
func HandleExternalAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Extract provider ID from URL path
	path := r.URL.Path
	providerID := path[len("/auth/callback/"):]

	// Verify state parameter to prevent CSRF
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		http.Error(w, "OAuth state mismatch", http.StatusBadRequest)
		return
	}

	// Get the original OAuth parameters
	clientIDCookie, err := r.Cookie("client_id")
	if err != nil {
		http.Error(w, "Missing client_id", http.StatusBadRequest)
		return
	}
	clientID := clientIDCookie.Value

	redirectURICookie, err := r.Cookie("redirect_uri")
	if err != nil {
		http.Error(w, "Missing redirect_uri", http.StatusBadRequest)
		return
	}
	redirectURI := redirectURICookie.Value

	// Get authorization code from query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Get provider configuration
	provider, err := repositories.GetAuthProviderByID(providerID)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Create the OAuth2 provider
	oauthProvider, err := adapters.GetProvider(provider)
	if err != nil {
		http.Error(w, "Failed to initialize provider", http.StatusInternalServerError)
		return
	}

	// Exchange code for access token
	token, err := oauthProvider.ExchangeCodeForToken(code, fmt.Sprintf("http://%s/auth/callback/%s", r.Host, providerID))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info from provider
	userInfo, err := oauthProvider.GetUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract user details
	var userID, username, email string

	// Directly access fields of oauthProvider since it is a struct
	if id, ok := userInfo[oauthProvider.UserIDField].(string); ok {
		userID = fmt.Sprintf("%s_%s", provider.Type, id)
	}
	if name, ok := userInfo[oauthProvider.UserNameField].(string); ok {
		username = name
	}
	if mail, ok := userInfo[oauthProvider.UserEmailField].(string); ok {
		email = mail
	}

	if userID == "" || username == "" {
		http.Error(w, "Failed to extract user information", http.StatusInternalServerError)
		return
	}

	if provider.Type == models.GitHubProvider && email == "" {
		// Check if the provider type is GitHub and email is missing
		if provider.Type == models.GitHubProvider && email == "" {
			// Attempt to fetch email from GitHub API
			githubEmail, err := adapters.GetUserEmailFromGitHub(token)
			if err == nil && githubEmail != "" {
				email = githubEmail
			}
		}
	}

	var user *models.User
	if config.App.UserProvider.Type == "external" {
		user, err = userAdapters.CurrentUserProvider.GetUserByEmail(email)
		if err != nil {
			http.Error(w, "Failed to get user account", http.StatusInternalServerError)
			return
		}
	} else {
		user, err = repositories.GetUserByExternalID(userID)
		if err != nil {
			// User doesn't exist, create a new one
			user, err = repositories.CreateExternalUser(userID, username, email, string(provider.Type))
			if err != nil {
				http.Error(w, "Failed to create user account", http.StatusInternalServerError)
				return
			}
		}
	}

	// Generate an auth code for the OAuth flow
	authCode := uuid.NewString()
	err = repositories.StoreAuthCode(&models.AuthCode{
		Code:        authCode,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		UserID:      user.ID,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		Scope:       "openid profile email",
	})
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "client_id", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "redirect_uri", MaxAge: -1, Path: "/"})

	// Redirect back to the client with the auth code
	http.Redirect(w, r, redirectURI+"?code="+authCode, http.StatusFound)
}
