package adapters

import (
	"zenauth/internal/models"
)

type UserProvider interface {
	GetUserByUsername(username string) (*models.User, error)

	GetUserByEmail(email string) (*models.User, error)

	VerifyPassword(hashedPassword, password string) bool

	GetAllUsers() ([]models.User, error)
}

type ExternalUserCreator interface {
	GetUserByExternalID(externalID string) (*models.User, error)

	CreateExternalUser(externalID string, username string, email string, provider string) (*models.User, error)
}
