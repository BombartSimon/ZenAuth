// Package router provides HTTP routing functionality for ZenAuth
package router

import (
	"net/http"

	"zenauth/internal/handlers"
	"zenauth/internal/middlewares"
	"zenauth/internal/oauth"

	"github.com/gorilla/mux"
)

// Router encapsulates the mux router and provides helper methods
type Router struct {
	root   *mux.Router
	admin  *mux.Router
	api    *mux.Router
	public *mux.Router
}

// New creates a new Router with all routes configured
func New(oauthFlows []oauth.OAuthFlow) *Router {
	handlers.RegisterFlows(oauthFlows)

	r := &Router{
		root: mux.NewRouter(),
	}

	// Setup subrouters
	r.api = r.root.PathPrefix("/api").Subrouter()
	r.public = r.root.PathPrefix("").Subrouter()

	// Admin subrouter with auth middleware
	adminBase := r.root.PathPrefix("/admin").Subrouter()

	// Login routes that don't require authentication
	adminLoginRouter := adminBase.PathPrefix("").Subrouter()
	adminLoginRouter.HandleFunc("/login", handlers.AdminLoginPageHandler).Methods("GET")
	adminLoginRouter.HandleFunc("/login/submit", handlers.AdminLoginHandler).Methods("POST")
	// adminLoginRouter.HandleFunc("/logout", handlers.AdminLogoutHandler).Methods("POST")  // Uncomment when implemented

	// Protected admin routes with auth middleware
	r.admin = adminBase.PathPrefix("").Subrouter()
	r.admin.Use(middlewares.AdminAuthMiddleware)

	r.setupRoutes()
	return r
}

// Handler returns the HTTP handler for the router
func (r *Router) Handler() http.Handler {
	return r.root
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	r.setupPublicRoutes()
	r.setupAdminRoutes()
	r.setupAPIRoutes()
}

// setupPublicRoutes configures public-facing routes
func (r *Router) setupPublicRoutes() {
	// OAuth endpoints
	r.public.HandleFunc("/authorize", handlers.AuthorizeHandler).Methods("GET", "POST")
	r.public.Handle("/token", middlewares.WithCORS(http.HandlerFunc(handlers.TokenHandler))).Methods("POST")
	r.public.Handle("/userinfo", middlewares.WithCORS(http.HandlerFunc(handlers.UserInfoHandler))).Methods("GET")

	// External auth endpoints
	r.public.HandleFunc("/auth/external", handlers.StartExternalAuth).Methods("GET")
	r.public.HandleFunc("/auth/callback/{provider}", handlers.HandleExternalAuthCallback).Methods("GET")

	// Static files
	r.public.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}

// setupAdminRoutes configures admin console routes
func (r *Router) setupAdminRoutes() {
	// Admin static files (protected)
	r.admin.PathPrefix("/assets/").Handler(http.StripPrefix("/admin/assets/", http.FileServer(http.Dir("./assets"))))

	// Handle JS files with proper MIME type
	r.admin.PathPrefix("/js/").HandlerFunc(handlers.StaticFileHandler)

	// Handle CSS files with proper MIME type
	r.admin.HandleFunc("/styles.css", handlers.StaticFileHandler)

	// Index and other files
	r.admin.HandleFunc("/", handlers.StaticFileHandler)

	// User management
	r.admin.HandleFunc("/users", handlers.AdminUsersHandler).Methods("GET", "POST")
	r.admin.HandleFunc("/users/{id}", handlers.AdminUserHandler).Methods("GET", "PUT", "DELETE")
	r.admin.HandleFunc("/blocked-users", handlers.AdminBlockedUsersHandler).Methods("GET")
	r.admin.HandleFunc("/unblock-user", handlers.AdminUnblockUserHandler).Methods("POST")

	// Client management
	r.admin.HandleFunc("/clients", handlers.AdminClientsHandler).Methods("GET", "POST")
	r.admin.HandleFunc("/clients/{id}", handlers.AdminClientHandler).Methods("GET", "PUT", "DELETE")

	// Role management
	r.admin.HandleFunc("/roles", handlers.AdminRolesHandler).Methods("GET", "POST")
	r.admin.HandleFunc("/roles/{id}", handlers.AdminRoleHandler).Methods("GET", "PUT", "DELETE")

	// Group management
	r.admin.HandleFunc("/groups", handlers.AdminGroupsHandler).Methods("GET", "POST")
	r.admin.HandleFunc("/groups/{id}", handlers.AdminGroupHandler).Methods("GET", "PUT", "DELETE")

	// User-role and user-group assignments
	r.admin.HandleFunc("/users-roles", handlers.AdminUserRolesHandler).Methods("GET", "POST", "DELETE")
	r.admin.HandleFunc("/users-groups", handlers.AdminUserGroupsHandler).Methods("GET", "POST", "DELETE")

	// Auth providers
	r.admin.HandleFunc("/auth-providers", handlers.ListAuthProviders).Methods("GET")
	r.admin.HandleFunc("/auth-providers", handlers.CreateAuthProvider).Methods("POST")
	r.admin.HandleFunc("/auth-providers/{id}", handlers.GetAuthProvider).Methods("GET")
	r.admin.HandleFunc("/auth-providers/{id}", handlers.UpdateAuthProvider).Methods("PUT")
	r.admin.HandleFunc("/auth-providers/{id}", handlers.DeleteAuthProvider).Methods("DELETE")
}

// setupAPIRoutes configures API endpoints
func (r *Router) setupAPIRoutes() {
	// Add API routes here when needed
}
