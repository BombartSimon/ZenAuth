package adapters

import (
	"encoding/json"
	"net/http"
	"zenauth/internal/oauth"
)

func NewGitHubProvider(clientID, clientSecret string) oauth.OAuth2Provider {
	return oauth.OAuth2Provider{
		Type:                 oauth.GitHub,
		ClientID:             clientID,
		ClientSecret:         clientSecret,
		AuthURL:              "https://github.com/login/oauth/authorize",
		TokenURL:             "https://github.com/login/oauth/access_token",
		UserInfoURL:          "https://api.github.com/user",
		Scopes:               []string{"read:user", "user:email"},
		UserIDField:          "id",
		UserNameField:        "login",
		UserEmailField:       "email",
		AdditionalAuthParams: map[string]string{},
	}
}

func GetUserEmailFromGitHub(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Search for the primary verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// If no primary email is found, return the first verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", nil
}
