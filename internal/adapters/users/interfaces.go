package adapters

import (
	"zenauth/internal/models"
)

// UserProvider defines the interface for retrieving user information
type UserProvider interface {
	// GetUserByUsername retrieves a user by their username
	GetUserByUsername(username string) (*models.User, error)

	// GetUserByEmail retrieves a user by their email
	GetUserByEmail(email string) (*models.User, error)

	// VerifyPassword checks if the provided password matches the user's password
	VerifyPassword(hashedPassword, password string) bool

	// GetUserByID retrieves a user by their ID
	// GetUserByID(id string) (*models.User, error)

	// GetAllUsers returns all users from the provider
	GetAllUsers() ([]models.User, error)
}
