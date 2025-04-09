package repositories

import (
	"errors"
	"zenauth/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// User Management

// GetAllUsers returns a list of all users in the database
func GetAllUsers() ([]models.User, error) {
	rows, err := db.Query("SELECT id, username, password_hash FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// CreateUser creates a new user with the provided username and password
func CreateUser(username, password string) (*models.User, error) {
	// Hash the password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create a new UUID
	id := uuid.New().String()

	// Insert the user into the database
	_, err = db.Exec("INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)",
		id, username, passwordHash)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           id,
		Username:     username,
		PasswordHash: string(passwordHash),
	}, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id string) (*models.User, error) {
	row := db.QueryRow("SELECT id, username, password_hash FROM users WHERE id = $1", id)
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user's username and/or password
func UpdateUser(id, username string, newPassword *string) error {
	if newPassword == nil {
		// Only update the username
		_, err := db.Exec("UPDATE users SET username = $1 WHERE id = $2", username, id)
		return err
	}

	// Hash the new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update both username and password
	_, err = db.Exec("UPDATE users SET username = $1, password_hash = $2 WHERE id = $3",
		username, passwordHash, id)
	return err
}

// DeleteUser deletes a user by ID
func DeleteUser(id string) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// Client Management

// GetAllClients returns a list of all OAuth clients
func GetAllClients() ([]models.Client, error) {
	rows, err := db.Query("SELECT id, secret, name, redirect_uris FROM clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var client models.Client
		if err := rows.Scan(&client.ID, &client.Secret, &client.Name, pq.Array(&client.RedirectURIs)); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}

// CreateClient creates a new OAuth client
func CreateClient(id, secret, name string, redirectURIs []string) (*models.Client, error) {
	// Check if client ID already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM clients WHERE id = $1", id).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("client ID already exists")
	}

	// Insert the client
	_, err = db.Exec("INSERT INTO clients (id, secret, name, redirect_uris) VALUES ($1, $2, $3, $4)",
		id, secret, name, pq.Array(redirectURIs))
	if err != nil {
		return nil, err
	}

	return &models.Client{
		ID:           id,
		Secret:       secret,
		Name:         name,
		RedirectURIs: redirectURIs,
	}, nil
}

// UpdateClient updates an OAuth client
func UpdateClient(id, name string, secret *string, redirectURIs []string) error {
	if secret == nil {
		// Only update name and redirect URIs
		_, err := db.Exec("UPDATE clients SET name = $1, redirect_uris = $2 WHERE id = $3",
			name, pq.Array(redirectURIs), id)
		return err
	}

	// Update name, secret, and redirect URIs
	_, err := db.Exec("UPDATE clients SET name = $1, secret = $2, redirect_uris = $3 WHERE id = $4",
		name, *secret, pq.Array(redirectURIs), id)
	return err
}

// DeleteClient deletes an OAuth client by ID
func DeleteClient(id string) error {
	// First delete related refresh tokens
	_, err := db.Exec("DELETE FROM refresh_tokens WHERE client_id = $1", id)
	if err != nil {
		return err
	}

	// Then delete related auth codes
	_, err = db.Exec("DELETE FROM auth_codes WHERE client_id = $1", id)
	if err != nil {
		return err
	}

	// Finally delete the client itself
	_, err = db.Exec("DELETE FROM clients WHERE id = $1", id)
	return err
}
