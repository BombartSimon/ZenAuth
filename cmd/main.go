package main

import (
	"log"
	"net/http"
	"zenauth/config"
	"zenauth/internal/handlers"
	"zenauth/internal/middlewares"
	"zenauth/internal/oauth"
	"zenauth/internal/repositories"

	rProviders "zenauth/internal/adapters/role"
	uProviders "zenauth/internal/adapters/users"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load(".env")

	config.Load()

	err := repositories.InitPostgres("postgres://oauth_user:oauth_pass@localhost:5432/oauth?sslmode=disable")
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// Initialiser le provider d'utilisateurs
	if err := uProviders.InitUserProvider(); err != nil {
		log.Fatalf("❌ Failed to initialize user provider: %v", err)
	}
	log.Println("✅ User provider initialized")

	// Initialiser le gestionnaire de rôles
	if err := rProviders.InitRoleManager(); err != nil {
		log.Fatalf("❌ Failed to initialize role manager: %v", err)
	}
	log.Println("✅ Role manager initialized")

	flows := []oauth.OAuthFlow{
		&oauth.ClientCredentialsFlow{},
		&oauth.RefreshTokenFlow{},
		&oauth.AuthorizationCodeFlow{},
	}
	handlers.RegisterFlows(flows)

	// OAuth endpoints
	http.HandleFunc("/authorize", handlers.AuthorizeHandler)
	http.Handle("/token", middlewares.WithCORS(http.HandlerFunc(handlers.TokenHandler)))
	http.Handle("/userinfo", middlewares.WithCORS(http.HandlerFunc(handlers.UserInfoHandler)))

	// Admin endpoints - Users
	http.HandleFunc("/admin/users", handlers.AdminUsersHandler)
	http.HandleFunc("/admin/users/", handlers.AdminUserHandler)

	// Admin endpoints - Clients
	http.HandleFunc("/admin/clients", handlers.AdminClientsHandler)
	http.HandleFunc("/admin/clients/", handlers.AdminClientHandler)

	// Admin endpoints - Roles (nouveaux endpoints)
	http.HandleFunc("/admin/roles", handlers.AdminRolesHandler)
	http.HandleFunc("/admin/roles/", handlers.AdminRoleHandler)

	// Admin endpoints - Groups (nouveaux endpoints)
	http.HandleFunc("/admin/groups", handlers.AdminGroupsHandler)
	http.HandleFunc("/admin/groups/", handlers.AdminGroupHandler)

	// Admin endpoints - User Roles (nouveaux endpoints)
	http.HandleFunc("/admin/users-roles", handlers.AdminUserRolesHandler)
	http.HandleFunc("/admin/users-groups", handlers.AdminUserGroupsHandler)

	// Static files using custom handler
	http.HandleFunc("/admin/", handlers.StaticFileHandler)

	log.Println("🚀 OAuth server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
