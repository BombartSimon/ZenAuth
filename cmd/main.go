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
	sProviders "zenauth/internal/adapters/sessions"
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

	// Initialiser le gestionnaire de sessions
	if err := sProviders.InitSessions(); err != nil {
		log.Fatalf("❌ Failed to initialize session manager: %v", err)
	}
	log.Println("✅ Session manager initialized")

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

	http.HandleFunc("/admin/blocked-users", handlers.AdminBlockedUsersHandler)
	http.HandleFunc("/admin/unblock-user", handlers.AdminUnblockUserHandler)
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

	// Static files handler for /static/ path
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// Static files using custom handler
	http.HandleFunc("/admin/", handlers.StaticFileHandler)

	// Auth Provider Admin API
	http.HandleFunc("/admin/auth-providers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListAuthProviders(w, r)
		case http.MethodPost:
			handlers.CreateAuthProvider(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Auth Provider Admin API - Individual provider operations
	http.HandleFunc("/admin/auth-providers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetAuthProvider(w, r)
		case http.MethodPut:
			handlers.UpdateAuthProvider(w, r)
		case http.MethodDelete:
			handlers.DeleteAuthProvider(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/auth/external", handlers.StartExternalAuth)
	http.HandleFunc("/auth/callback/", handlers.HandleExternalAuthCallback)

	log.Println("🚀 OAuth server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
