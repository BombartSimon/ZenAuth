package models

type AuthProviderType string

const (
	GoogleProvider    AuthProviderType = "google"
	MicrosoftProvider AuthProviderType = "microsoft"
	GitHubProvider    AuthProviderType = "github"
)

type AuthProvider struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Type         AuthProviderType `json:"type"`
	ClientID     string           `json:"client_id"`
	ClientSecret string           `json:"-"`         // Don't expose in JSON responses
	TenantID     string           `json:"tenant_id"` // For Microsoft providers
	Enabled      bool             `json:"enabled"`
	CreatedAt    string           `json:"created_at"`
}
