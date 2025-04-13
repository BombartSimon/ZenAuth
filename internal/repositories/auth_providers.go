package repositories

import (
	"time"
	"zenauth/internal/models"

	"github.com/google/uuid"
)

// GetAllAuthProviders returns all configured authentication providers
func GetAllAuthProviders() ([]models.AuthProvider, error) {
	rows, err := db.Query("SELECT id, name, type, client_id, client_secret, tenant_id, enabled, created_at FROM auth_providers ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.AuthProvider
	for rows.Next() {
		var provider models.AuthProvider
		var createdAt time.Time
		if err := rows.Scan(&provider.ID, &provider.Name, &provider.Type,
			&provider.ClientID, &provider.ClientSecret,
			&provider.TenantID, &provider.Enabled, &createdAt); err != nil {
			return nil, err
		}
		provider.CreatedAt = createdAt.Format(time.RFC3339)
		providers = append(providers, provider)
	}

	return providers, nil
}

// GetEnabledAuthProviders returns only enabled authentication providers
func GetEnabledAuthProviders() ([]models.AuthProvider, error) {
	rows, err := db.Query("SELECT id, name, type, client_id, client_secret, tenant_id, enabled, created_at FROM auth_providers WHERE enabled = true ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.AuthProvider
	for rows.Next() {
		var provider models.AuthProvider
		var createdAt time.Time
		if err := rows.Scan(&provider.ID, &provider.Name, &provider.Type,
			&provider.ClientID, &provider.ClientSecret,
			&provider.TenantID, &provider.Enabled, &createdAt); err != nil {
			return nil, err
		}
		provider.CreatedAt = createdAt.Format(time.RFC3339)
		providers = append(providers, provider)
	}

	return providers, nil
}

// GetAuthProviderByID retrieves a provider by ID
func GetAuthProviderByID(id string) (*models.AuthProvider, error) {
	var provider models.AuthProvider
	var createdAt time.Time
	err := db.QueryRow("SELECT id, name, type, client_id, client_secret, tenant_id, enabled, created_at FROM auth_providers WHERE id = $1", id).
		Scan(&provider.ID, &provider.Name, &provider.Type, &provider.ClientID,
			&provider.ClientSecret, &provider.TenantID, &provider.Enabled, &createdAt)
	if err != nil {
		return nil, err
	}
	provider.CreatedAt = createdAt.Format(time.RFC3339)
	return &provider, nil
}

// CreateAuthProvider creates a new authentication provider
func CreateAuthProvider(name string, providerType models.AuthProviderType,
	clientID, clientSecret string, tenantID string) (*models.AuthProvider, error) {
	id := uuid.NewString()
	now := time.Now()

	_, err := db.Exec("INSERT INTO auth_providers (id, name, type, client_id, client_secret, tenant_id, enabled, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		id, name, providerType, clientID, clientSecret, tenantID, false, now)
	if err != nil {
		return nil, err
	}

	return &models.AuthProvider{
		ID:           id,
		Name:         name,
		Type:         providerType,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     tenantID,
		Enabled:      false,
		CreatedAt:    now.Format(time.RFC3339),
	}, nil
}

// UpdateAuthProvider updates an existing authentication provider
func UpdateAuthProvider(id, name string, clientID, clientSecret, tenantID string, enabled bool) error {
	if clientSecret == "" {
		// Don't update the secret if not provided
		_, err := db.Exec("UPDATE auth_providers SET name = $1, client_id = $2, tenant_id = $3, enabled = $4 WHERE id = $5",
			name, clientID, tenantID, enabled, id)
		return err
	}

	// Update including the secret
	_, err := db.Exec("UPDATE auth_providers SET name = $1, client_id = $2, client_secret = $3, tenant_id = $4, enabled = $5 WHERE id = $6",
		name, clientID, clientSecret, tenantID, enabled, id)
	return err
}

// DeleteAuthProvider deletes an authentication provider
func DeleteAuthProvider(id string) error {
	_, err := db.Exec("DELETE FROM auth_providers WHERE id = $1", id)
	return err
}

// GetUserByExternalID retrieves a user by their external provider ID
func GetUserByExternalID(externalID string) (*models.User, error) {
	var user models.User
	err := db.QueryRow("SELECT id, username, password_hash, email FROM users WHERE external_id = $1", externalID).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email)
	return &user, err
}

// CreateExternalUser creates a new user from an external provider
func CreateExternalUser(externalID, username, email, provider string) (*models.User, error) {
	// Generate a UUID for the user
	id := uuid.NewString()

	// Create empty password hash for external users
	_, err := db.Exec("INSERT INTO users (id, username, password_hash, email, external_id, auth_provider) VALUES ($1, $2, '', $3, $4, $5)",
		id, username, email, externalID, provider)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}
