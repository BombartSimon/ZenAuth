package adapters

import (
	"fmt"
	"zenauth/internal/oauth"
)

func NewMicrosoftProvider(clientID, clientSecret, tenantID string) oauth.OAuth2Provider {
	if tenantID == "" {
		tenantID = "common"
	}

	return oauth.OAuth2Provider{
		Type:           oauth.Microsoft,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		AuthURL:        fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", tenantID),
		TokenURL:       fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID),
		UserInfoURL:    "https://graph.microsoft.com/v1.0/me",
		Scopes:         []string{"openid", "email", "profile", "User.Read"},
		UserIDField:    "id",
		UserNameField:  "displayName",
		UserEmailField: "mail",
	}
}
