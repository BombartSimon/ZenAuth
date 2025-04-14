package main

import (
	"log"
	"net/http"
	"zenauth/config"
	"zenauth/internal/oauth"
	"zenauth/internal/repositories"
	"zenauth/internal/router"

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

	// Configuration des flux OAuth
	flows := []oauth.OAuthFlow{
		&oauth.ClientCredentialsFlow{},
		&oauth.RefreshTokenFlow{},
		&oauth.AuthorizationCodeFlow{},
	}

	// Création et configuration du routeur
	r := router.New(flows)

	log.Println("🚀 OAuth server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r.Handler()))
}
