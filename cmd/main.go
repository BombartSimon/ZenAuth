package main

import (
	"log"
	"net/http"
	"zenauth/config"
	"zenauth/handlers"
	"zenauth/middlewares"
	"zenauth/oauth"
	"zenauth/oauth/store"

	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load(".env")

	// 🔧 Chargement de la configuration
	config.Load()

	// 🔌 Connexion à PostgreSQL
	err := store.InitPostgres("postgres://oauth_user:oauth_pass@localhost:5432/oauth?sslmode=disable")
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// 💡 Enregistrement des flows OAuth
	flows := []oauth.OAuthFlow{
		&oauth.ClientCredentialsFlow{},
		&oauth.RefreshTokenFlow{},
		&oauth.AuthorizationCodeFlow{}, // ← tu rajouteras ce flow après
	}
	handlers.RegisterFlows(flows)

	// 📡 Routing
	http.HandleFunc("/authorize", handlers.AuthorizeHandler)
	http.Handle("/token", middlewares.WithCORS(http.HandlerFunc(handlers.TokenHandler)))
	http.Handle("/userinfo", middlewares.WithCORS(http.HandlerFunc(handlers.UserInfoHandler)))

	log.Println("🚀 OAuth server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
