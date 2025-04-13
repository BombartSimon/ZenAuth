package adapters

import (
	"database/sql"
	"errors"
	"zenauth/config"
)

func InitRoleManager() error {
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

func initLocalRoleManager() error {
	db, err := sql.Open("postgres", config.App.DatabaseURL)
	if err != nil {
		return err
	}

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

func initExternalRoleManager() error {
	db, err := sql.Open("postgres", config.App.RoleManager.ExternalConn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	config := ExternalRoleConfig{
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
