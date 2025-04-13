package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort     string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string

	Admin struct {
		JWTSecret string
	}

	UserProvider struct {
		Type string `json:"type"` // "sql", "rest"
		// SQL Options
		SQLConn       string `json:"sqlConn,omitempty"`
		SQLTable      string `json:"sqlTable,omitempty"`
		SQLIDField    string `json:"sqlIdField,omitempty"`
		SQLUserField  string `json:"sqlUserField,omitempty"`
		SQLPassField  string `json:"sqlPassField,omitempty"`
		SQLEmailField string `json:"sqlEmailField,omitempty"`

		// REST Options
		RESTURL  string `json:"restUrl,omitempty"`
		RESTAuth string `json:"restAuth,omitempty"`
	}

	// Add RoleManager struct
	RoleManager struct {
		// Managee type (local, external)
		Type string `json:"type"`

		ExternalConn string `json:"externalConn,omitempty"`

		RoleTable      string `json:"roleTable,omitempty"`
		GroupTable     string `json:"groupTable,omitempty"`
		UserRoleTable  string `json:"userRoleTable,omitempty"`
		GroupRoleTable string `json:"groupRoleTable,omitempty"`
		UserGroupTable string `json:"userGroupTable,omitempty"`

		RoleIDCol   string `json:"roleIdCol,omitempty"`
		RoleNameCol string `json:"roleNameCol,omitempty"`
		RoleDescCol string `json:"roleDescCol,omitempty"`

		GroupIDCol   string `json:"groupIdCol,omitempty"`
		GroupNameCol string `json:"groupNameCol,omitempty"`
		GroupDescCol string `json:"groupDescCol,omitempty"`

		UserRoleUserCol   string `json:"userRoleUserCol,omitempty"`
		UserRoleRoleCol   string `json:"userRoleRoleCol,omitempty"`
		GroupRoleGroupCol string `json:"groupRoleGroupCol,omitempty"`
		GroupRoleRoleCol  string `json:"groupRoleRoleCol,omitempty"`
		UserGroupUserCol  string `json:"userGroupUserCol,omitempty"`
		UserGroupGroupCol string `json:"userGroupGroupCol,omitempty"`

		IncludeRolesInJWT bool `json:"includeRolesInJWT,omitempty"`
	}

	// Rate limiting configuration
	RateLimit struct {
		Enabled           bool
		MaxAttempts       int
		BlockDuration     time.Duration
		CounterExpiration time.Duration
		Provider          string // "memcached", "redis", etc.
		ConnectionURL     string // "localhost:11211" for Memcached, "redis://..." pour Redis
	}
}

var App Config

func Load() {
	App = Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://oauth_user:oauth_pass@localhost:5432/oauth?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "supersecretkey"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
	}

	// Admin JWT secret
	App.Admin.JWTSecret = getEnv("ADMIN_JWT_SECRET", App.JWTSecret)

	// User provider configuration
	App.UserProvider.Type = getEnv("USER_PROVIDER_TYPE", "default")
	App.UserProvider.SQLConn = getEnv("USER_PROVIDER_SQL_CONN", App.DatabaseURL)
	App.UserProvider.SQLTable = getEnv("USER_PROVIDER_SQL_TABLE", "users")
	App.UserProvider.SQLIDField = getEnv("USER_PROVIDER_SQL_ID_FIELD", "id")
	App.UserProvider.SQLUserField = getEnv("USER_PROVIDER_SQL_USER_FIELD", "username")
	App.UserProvider.SQLPassField = getEnv("USER_PROVIDER_SQL_PASS_FIELD", "password_hash")
	App.UserProvider.SQLEmailField = getEnv("USER_PROVIDER_SQL_EMAIL_FIELD", "email")
	App.UserProvider.RESTURL = getEnv("USER_PROVIDER_REST_URL", "")
	App.UserProvider.RESTAuth = getEnv("USER_PROVIDER_REST_AUTH", "")

	// Role manager configuration
	App.RoleManager.Type = getEnv("ROLE_MANAGER_TYPE", "local")
	App.RoleManager.ExternalConn = getEnv("ROLE_MANAGER_EXTERNAL_CONN", App.UserProvider.SQLConn)
	App.RoleManager.IncludeRolesInJWT = getEnvBool("ROLE_MANAGER_INCLUDE_IN_JWT", true)

	// Table names and column names
	App.RoleManager.RoleTable = getEnv("ROLE_MANAGER_ROLE_TABLE", "roles")
	App.RoleManager.GroupTable = getEnv("ROLE_MANAGER_GROUP_TABLE", "groups")
	App.RoleManager.UserRoleTable = getEnv("ROLE_MANAGER_USER_ROLE_TABLE", "user_roles")
	App.RoleManager.GroupRoleTable = getEnv("ROLE_MANAGER_GROUP_ROLE_TABLE", "group_roles")
	App.RoleManager.UserGroupTable = getEnv("ROLE_MANAGER_USER_GROUP_TABLE", "user_groups")

	App.RoleManager.RoleIDCol = getEnv("ROLE_MANAGER_ROLE_ID_COL", "id")
	App.RoleManager.RoleNameCol = getEnv("ROLE_MANAGER_ROLE_NAME_COL", "name")
	App.RoleManager.RoleDescCol = getEnv("ROLE_MANAGER_ROLE_DESC_COL", "description")

	App.RoleManager.GroupIDCol = getEnv("ROLE_MANAGER_GROUP_ID_COL", "id")
	App.RoleManager.GroupNameCol = getEnv("ROLE_MANAGER_GROUP_NAME_COL", "name")
	App.RoleManager.GroupDescCol = getEnv("ROLE_MANAGER_GROUP_DESC_COL", "description")

	App.RoleManager.UserRoleUserCol = getEnv("ROLE_MANAGER_USER_ROLE_USER_COL", "user_id")
	App.RoleManager.UserRoleRoleCol = getEnv("ROLE_MANAGER_USER_ROLE_ROLE_COL", "role_id")
	App.RoleManager.GroupRoleGroupCol = getEnv("ROLE_MANAGER_GROUP_ROLE_GROUP_COL", "group_id")
	App.RoleManager.GroupRoleRoleCol = getEnv("ROLE_MANAGER_GROUP_ROLE_ROLE_COL", "role_id")
	App.RoleManager.UserGroupUserCol = getEnv("ROLE_MANAGER_USER_GROUP_USER_COL", "user_id")
	App.RoleManager.UserGroupGroupCol = getEnv("ROLE_MANAGER_USER_GROUP_GROUP_COL", "group_id")

	// Rate limiting configuration
	App.RateLimit.Enabled = getEnvBool("RATE_LIMIT_ENABLED", true)
	App.RateLimit.MaxAttempts = getEnvInt("RATE_LIMIT_MAX_ATTEMPTS", 5)
	blockMinutes := getEnvInt("RATE_LIMIT_BLOCK_MINUTES", 30)
	App.RateLimit.BlockDuration = time.Duration(blockMinutes) * time.Minute
	counterHours := getEnvInt("RATE_LIMIT_COUNTER_HOURS", 24)
	App.RateLimit.CounterExpiration = time.Duration(counterHours) * time.Hour
	App.RateLimit.Provider = getEnv("RATE_LIMIT_PROVIDER", "memcached")
	App.RateLimit.ConnectionURL = getEnv("RATE_LIMIT_CONNECTION_URL", "localhost:11211")

	log.Println("✅ Configuration loaded")
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
