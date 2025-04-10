package adapters

import (
	"database/sql"
	"errors"
	"log"
	"zenauth/config"
)

var (
	// CurrentUserProvider holds the active user provider
	CurrentUserProvider UserProvider
)

// InitUserProvider initializes the appropriate user provider based on configuration
func InitUserProvider() error {
	switch config.App.UserProvider.Type {
	case "external":
		return initSQLUserProvider()
	case "rest":
		return initRESTUserProvider()
	case "local":
		fallthrough
	default:
		return errors.New("unsupported user provider type: " + config.App.UserProvider.Type)
	}
}

// Initialize SQL User Provider
func initSQLUserProvider() error {
	db, err := sql.Open("postgres", config.App.UserProvider.SQLConn)
	if err != nil {
		return err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return err
	}

	CurrentUserProvider = NewSQLUser(
		db,
		config.App.UserProvider.SQLTable,
		config.App.UserProvider.SQLIDField,
		config.App.UserProvider.SQLUserField,
		config.App.UserProvider.SQLPassField,
		config.App.UserProvider.SQLEmailField,
	)

	log.Println("✅ SQL User Provider initialized with table:", config.App.UserProvider.SQLTable)
	return nil
}

// Initialize REST User Provider
func initRESTUserProvider() error {
	// This would be implemented in the future if needed
	return errors.New("REST User Provider not implemented yet")
}
