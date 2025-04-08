package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"
	"zenauth/models"
	"zenauth/oauth/store"

	"github.com/google/uuid"
)

var loginTmpl = template.Must(template.ParseFiles("templates/login.html.tmpl"))

func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		appName := r.URL.Query().Get("client_id")
		redirectURI := r.URL.Query().Get("redirect_uri")
		codeChallenge := r.URL.Query().Get("code_challenge")
		codeMethod := r.URL.Query().Get("code_challenge_method")
		logo := "/logo.png"
		loginTmpl.Execute(w, map[string]string{
			"ClientID":            appName,
			"RedirectURI":         redirectURI,
			"CodeChallenge":       codeChallenge,
			"CodeChallengeMethod": codeMethod,
			"Logo":                logo,
		})
		return
	}

	// POST (login form)
	username := r.FormValue("username")
	password := r.FormValue("password")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	codeChallenge := r.FormValue("code_challenge")
	codeMethod := r.FormValue("code_challenge_method")

	client, err := store.GetClientByID(clientID)
	if err != nil {
		http.Error(w, "unauthorized_client", http.StatusBadRequest)
		return
	}

	if !isRedirectURIAuthorized(redirectURI, client.RedirectURIs) {
		http.Error(w, "invalid_redirect_uri", http.StatusBadRequest)
		return
	}

	user, err := store.GetUserByUsername(username)
	if err != nil || !store.VerifyPassword(user.PasswordHash, password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	code := uuid.NewString()
	err = store.StoreAuthCode(&models.AuthCode{
		Code:                code,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		UserID:              user.ID,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
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
