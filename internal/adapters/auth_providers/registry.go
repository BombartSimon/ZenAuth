package adapters

import (
	"errors"
	"zenauth/internal/models"
	"zenauth/internal/oauth"
)

// GetProvider returns the appropriate OAuth2 provider based on the configuration
func GetProvider(provider *models.AuthProvider) (oauth.OAuth2Provider, error) {
	switch provider.Type {
	case models.GoogleProvider:
		return NewGoogleProvider(provider.ClientID, provider.ClientSecret), nil
	case models.MicrosoftProvider:
		return NewMicrosoftProvider(provider.ClientID, provider.ClientSecret, provider.TenantID), nil
	case models.GitHubProvider:
		return NewGitHubProvider(provider.ClientID, provider.ClientSecret), nil
	default:
		return oauth.OAuth2Provider{}, errors.New("unsupported provider type")
	}
}
