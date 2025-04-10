package adapters

import "zenauth/internal/oauth"

func NewGoogleProvider(clientID, clientSecret string) oauth.OAuth2Provider {
	return oauth.OAuth2Provider{
		Type:           oauth.Google,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		AuthURL:        "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:       "https://oauth2.googleapis.com/token",
		UserInfoURL:    "https://www.googleapis.com/oauth2/v3/userinfo",
		Scopes:         []string{"openid", "email", "profile"},
		UserIDField:    "sub",
		UserNameField:  "name",
		UserEmailField: "email",
		AdditionalAuthParams: map[string]string{
			"access_type": "offline",
		},
	}
}
