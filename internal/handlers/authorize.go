package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
	adapters "zenauth/internal/adapters/users"
	"zenauth/internal/models"
	"zenauth/internal/repositories"

	"github.com/google/uuid"
)

var loginTmpl = template.Must(template.ParseFiles("templates/login.html.tmpl"))

func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		appName := r.URL.Query().Get("client_id")
		redirectURI := r.URL.Query().Get("redirect_uri")
		codeChallenge := r.URL.Query().Get("code_challenge")
		codeMethod := r.URL.Query().Get("code_challenge_method")
		scopes := r.URL.Query().Get("scope")
		logo := "/logo.png"
		loginTmpl.Execute(w, map[string]string{
			"ClientID":            appName,
			"RedirectURI":         redirectURI,
			"CodeChallenge":       codeChallenge,
			"CodeChallengeMethod": codeMethod,
			"Logo":                logo,
			"Scope":               scopes,
		})
		return
	}

	// POST (login form)
	identifier := r.FormValue("identifier")
	password := r.FormValue("password")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	codeChallenge := r.FormValue("code_challenge")
	codeMethod := r.FormValue("code_challenge_method")
	scope := r.FormValue("scope")

	client, err := repositories.GetClientByID(clientID)
	if err != nil {
		http.Error(w, "unauthorized_client", http.StatusBadRequest)
		return
	}

	if !isRedirectURIAuthorized(redirectURI, client.RedirectURIs) {
		http.Error(w, "invalid_redirect_uri", http.StatusBadRequest)
		return
	}

	var user *models.User
	var userErr error

	// Get user from the user adapters
	if strings.Contains(identifier, "@") {
		// If username contains '@', we assume it's an email
		user, userErr = adapters.CurrentUserProvider.GetUserByEmail(identifier)
		if userErr != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	} else {
		// Otherwise, we assume it's a username
		user, userErr = adapters.CurrentUserProvider.GetUserByUsername(identifier)
		if userErr != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !adapters.CurrentUserProvider.VerifyPassword(user.PasswordHash, password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	code := uuid.NewString()
	err = repositories.StoreAuthCode(&models.AuthCode{
		Code:                code,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		UserID:              user.ID,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		Scope:               scope,
	})
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Println("REDIRECT TO:", redirectURI+"?code="+code)

	http.Redirect(w, r, redirectURI+"?code="+code, http.StatusFound)
}

func isRedirectURIAuthorized(uri string, allowed []string) bool {
	for _, u := range allowed {
		if u == uri {
			return true
		}
	}
	return false
}
