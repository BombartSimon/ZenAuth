package role

import (
	"database/sql"
	"errors"
	"zenauth/config"
)

// InitRoleManager initialise le gestionnaire de rôles approprié
func InitRoleManager() error {
	// Déterminer quel type de gestionnaire utiliser
	roleManagerType := config.App.RoleManager.Type

	switch roleManagerType {
	case "local", "default", "":
		return initLocalRoleManager()
	case "external":
		return initExternalRoleManager()
	default:
		return errors.New("unsupported role manager type: " + roleManagerType)
	}
}

// initLocalRoleManager initialise un gestionnaire local de rôles
func initLocalRoleManager() error {
	// Utiliser la même base de données que ZenAuth
	db, err := sql.Open("postgres", config.App.DatabaseURL)
	if err != nil {
		return err
	}

	// Vérifier la connexion
	if err := db.Ping(); err != nil {
		return err
	}

	manager, err := NewLocalManager(db)
	if err != nil {
		return err
	}

	CurrentManager = manager
	return nil
}

// initExternalRoleManager initialise un gestionnaire externe de rôles
func initExternalRoleManager() error {
	// Se connecter à la base de données externe
	db, err := sql.Open("postgres", config.App.RoleManager.ExternalConn)
	if err != nil {
		return err
	}

	// Vérifier la connexion
	if err := db.Ping(); err != nil {
		return err
	}

	// Configuration des tables et colonnes pour le manager externe
	config := ExternalManagerConfig{
		RoleTable:      config.App.RoleManager.RoleTable,
		GroupTable:     config.App.RoleManager.GroupTable,
		UserRoleTable:  config.App.RoleManager.UserRoleTable,
		GroupRoleTable: config.App.RoleManager.GroupRoleTable,
		UserGroupTable: config.App.RoleManager.UserGroupTable,

		RoleIDCol:   config.App.RoleManager.RoleIDCol,
		RoleNameCol: config.App.RoleManager.RoleNameCol,
		RoleDescCol: config.App.RoleManager.RoleDescCol,

		GroupIDCol:   config.App.RoleManager.GroupIDCol,
		GroupNameCol: config.App.RoleManager.GroupNameCol,
		GroupDescCol: config.App.RoleManager.GroupDescCol,

		UserRoleUserCol:   config.App.RoleManager.UserRoleUserCol,
		UserRoleRoleCol:   config.App.RoleManager.UserRoleRoleCol,
		GroupRoleGroupCol: config.App.RoleManager.GroupRoleGroupCol,
		GroupRoleRoleCol:  config.App.RoleManager.GroupRoleRoleCol,
		UserGroupUserCol:  config.App.RoleManager.UserGroupUserCol,
		UserGroupGroupCol: config.App.RoleManager.UserGroupGroupCol,
	}

	manager := NewExternalManager(db, config)
	CurrentManager = manager
	return nil
}
