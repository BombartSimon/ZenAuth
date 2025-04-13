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

	// Main mux for all routes
	mux := http.NewServeMux()

	// OAuth endpoints
	mux.HandleFunc("/authorize", handlers.AuthorizeHandler)
	mux.Handle("/token", middlewares.WithCORS(http.HandlerFunc(handlers.TokenHandler)))
	mux.Handle("/userinfo", middlewares.WithCORS(http.HandlerFunc(handlers.UserInfoHandler)))

	// Create a separate mux for admin routes that will be protected
	adminMux := http.NewServeMux()

	// Admin routes - Users
	adminMux.HandleFunc("/admin/users", handlers.AdminUsersHandler)
	adminMux.HandleFunc("/admin/users/", handlers.AdminUserHandler)
	adminMux.HandleFunc("/admin/blocked-users", handlers.AdminBlockedUsersHandler)
	adminMux.HandleFunc("/admin/unblock-user", handlers.AdminUnblockUserHandler)

	// Admin routes - Clients
	adminMux.HandleFunc("/admin/clients", handlers.AdminClientsHandler)
	adminMux.HandleFunc("/admin/clients/", handlers.AdminClientHandler)

	// Admin routes - Roles
	adminMux.HandleFunc("/admin/roles", handlers.AdminRolesHandler)
	adminMux.HandleFunc("/admin/roles/", handlers.AdminRoleHandler)

	// Admin routes - Groups
	adminMux.HandleFunc("/admin/groups", handlers.AdminGroupsHandler)
	adminMux.HandleFunc("/admin/groups/", handlers.AdminGroupHandler)

	// Admin routes - User Roles and Groups
	adminMux.HandleFunc("/admin/users-roles", handlers.AdminUserRolesHandler)
	adminMux.HandleFunc("/admin/users-groups", handlers.AdminUserGroupsHandler)

	// Admin API - Auth Providers
	adminMux.HandleFunc("/admin/auth-providers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ListAuthProviders(w, r)
		case http.MethodPost:
			handlers.CreateAuthProvider(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	adminMux.HandleFunc("/admin/auth-providers/", func(w http.ResponseWriter, r *http.Request) {
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

	// Admin static file handling
	adminMux.HandleFunc("/admin/", handlers.StaticFileHandler)

	// Add the protected admin routes to the main mux, wrapped with the auth middleware
	mux.Handle("/admin/", middlewares.AdminAuthMiddleware(adminMux))

	// Exclude login pages from auth middleware
	mux.HandleFunc("/admin/login", handlers.AdminLoginPageHandler)
	mux.HandleFunc("/admin/login/submit", handlers.AdminLoginHandler)
	// mux.HandleFunc("/admin/logout", handlers.AdminLogoutHandler)

	// Static files handler for /static/ path (unprotected)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// External auth routes
	mux.HandleFunc("/auth/external", handlers.StartExternalAuth)
	mux.HandleFunc("/auth/callback/", handlers.HandleExternalAuthCallback)

	log.Println("🚀 OAuth server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
