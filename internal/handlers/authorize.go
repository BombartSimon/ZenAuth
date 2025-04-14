package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
	sessionsAdapters "zenauth/internal/adapters/sessions"
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

		// Get enabled external providers
		externalProviders, err := repositories.GetEnabledAuthProviders()
		if err != nil {
			log.Printf("Failed to get external providers: %v", err)
			// Continue without external providers
			externalProviders = []models.AuthProvider{}
		}

		loginTmpl.Execute(w, map[string]interface{}{
			"ClientID":            appName,
			"RedirectURI":         redirectURI,
			"CodeChallenge":       codeChallenge,
			"CodeChallengeMethod": codeMethod,
			"Logo":                logo,
			"Scope":               scopes,
			"ExternalProviders":   externalProviders,
		})
		return
	}

	if r.Method == http.MethodPost {
		identifier := r.FormValue("identifier")
		password := r.FormValue("password")
		redirectURI := r.FormValue("redirect_uri")
		clientID := r.FormValue("client_id")
		codeChallenge := r.FormValue("code_challenge")
		codeMethod := r.FormValue("code_challenge_method")
		scope := r.FormValue("scope")
		state := r.FormValue("state")

		ipAddress := getClientIP(r)
		blocked, message, err := sessionsAdapters.CheckRateLimit(ipAddress)
		if err != nil {
			log.Printf("Rate limiting error: %v", err)
		} else if blocked {
			data := loginData(clientID, redirectURI, codeChallenge, codeMethod, scope, state)
			data["Error"] = message
			loginTmpl.Execute(w, data)
			return
		}

		if identifier != "" {
			userKey := "user:" + identifier
			blocked, message, err := sessionsAdapters.CheckRateLimit(userKey)
			if err != nil {
				log.Printf("Rate limiting error: %v", err)
			} else if blocked {
				data := loginData(clientID, redirectURI, codeChallenge, codeMethod, scope, state)
				data["Error"] = message
				loginTmpl.Execute(w, data)
				return
			}
		}

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
		} else {
			// Otherwise, we assume it's a username
			user, userErr = adapters.CurrentUserProvider.GetUserByUsername(identifier)
		}

		// Authentication failures handling with rate limiting
		if userErr != nil || user == nil || !adapters.CurrentUserProvider.VerifyPassword(user.PasswordHash, password) {
			attempts, err := sessionsAdapters.RecordFailedLoginAttempt(ipAddress)
			if err != nil {
				log.Printf("Error recording failed attempt for IP %s: %v", ipAddress, err)
			} else {
				log.Printf("Failed login attempt from IP %s: %d attempts", ipAddress, attempts)
			}

			if identifier != "" {
				userKey := "user:" + identifier
				userAttempts, err := sessionsAdapters.RecordFailedLoginAttempt(userKey)
				if err != nil {
					log.Printf("Error recording failed attempt for user %s: %v", identifier, err)
				} else {
					log.Printf("Failed login attempt for user '%s': %d attempts", identifier, userAttempts)
				}

				if err := sessionsAdapters.CurrentLimiter.RecordUserIP(identifier, ipAddress); err != nil {
					log.Printf("Error recording user-IP association on failed attempt: %v", err)
				} else {
					log.Printf("Associated user '%s' with IP '%s' on failed login attempt", identifier, ipAddress)
				}
			}

			data := loginData(clientID, redirectURI, codeChallenge, codeMethod, scope, state)
			data["Error"] = "Invalid username or password"
			loginTmpl.Execute(w, data)
			return
		}

		// Successful authentication - reset rate limiting
		if err := sessionsAdapters.ResetLoginAttempts(ipAddress); err != nil {
			log.Printf("Error resetting rate limit for IP %s: %v", ipAddress, err)
		}

		if identifier != "" {
			userKey := "user:" + identifier
			if err := sessionsAdapters.ResetLoginAttempts(userKey); err != nil {
				log.Printf("Error resetting rate limit for user %s: %v", identifier, err)
			}

			if err := sessionsAdapters.CurrentLimiter.RecordUserIP(identifier, ipAddress); err != nil {
				log.Printf("Error recording user-IP association: %v", err)
			} else {
				log.Printf("Associated user '%s' with IP '%s' for rate limiting purposes", identifier, ipAddress)
			}
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

		// Add state to redirect if present
		redirectURL := redirectURI + "?code=" + code
		if state != "" {
			redirectURL += "&state=" + state
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

func loginData(clientID, redirectURI, codeChallenge, codeMethod, scope, state string) map[string]interface{} {
	logo := "/logo.png"
	externalProviders, err := repositories.GetEnabledAuthProviders()
	if err != nil {
		log.Printf("Failed to get external providers: %v", err)
		externalProviders = []models.AuthProvider{}
	}

	return map[string]interface{}{
		"ClientID":            clientID,
		"RedirectURI":         redirectURI,
		"CodeChallenge":       codeChallenge,
		"CodeChallengeMethod": codeMethod,
		"Logo":                logo,
		"Scope":               scope,
		"State":               state,
		"ExternalProviders":   externalProviders,
	}
}

func getClientIP(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	return ip
}

func isRedirectURIAuthorized(uri string, allowed []string) bool {
	for _, u := range allowed {
		if u == uri {
			return true
		}
	}
	return false
}
