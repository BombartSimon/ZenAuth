package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ProviderType string

const (
	Google    ProviderType = "google"
	Microsoft ProviderType = "microsoft"
	GitHub    ProviderType = "github"
)

type Provider interface {
	GetAuthURL(state, redirectURI string) string
	ExchangeCodeForToken(code, redirectURI string) (string, error)
	GetUserInfo(token string) (map[string]interface{}, error)
}

// OAuth2Provider implements the common OAuth2 flow for different providers
type OAuth2Provider struct {
	Type                 ProviderType
	ClientID             string
	ClientSecret         string
	AuthURL              string
	TokenURL             string
	UserInfoURL          string
	Scopes               []string
	UserIDField          string
	UserNameField        string
	UserEmailField       string
	AdditionalAuthParams map[string]string
}

// GetAuthURL generates the OAuth authorization URL
func (p *OAuth2Provider) GetAuthURL(state, redirectURI string) string {
	u, _ := url.Parse(p.AuthURL)
	q := u.Query()
	q.Set("client_id", p.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("scope", strings.Join(p.Scopes, " "))

	// Add any provider-specific params
	for k, v := range p.AdditionalAuthParams {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// ExchangeCodeForToken exchanges the authorization code for a token
func (p *OAuth2Provider) ExchangeCodeForToken(code, redirectURI string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", p.ClientID)
	data.Set("client_secret", p.ClientSecret)

	req, err := http.NewRequest("POST", p.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Spécifique à GitHub: accepter JSON comme réponse
	if p.Type == GitHub {
		req.Header.Add("Accept", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to exchange token: %s, %s", resp.Status, string(body))
	}

	var result map[string]interface{}

	// GitHub peut renvoyer des réponses au format différent
	if p.Type == GitHub && resp.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		values, err := url.ParseQuery(string(body))
		if err != nil {
			return "", err
		}

		result = make(map[string]interface{})
		result["access_token"] = values.Get("access_token")
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", errors.New("access_token not found in response")
	}

	return token, nil
}

// GetUserInfo fetches user information using the access token
func (p *OAuth2Provider) GetUserInfo(token string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", p.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	// Spécifique à GitHub
	if p.Type == GitHub {
		req.Header.Add("Accept", "application/vnd.github.v3+json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s, %s", resp.Status, string(body))
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
