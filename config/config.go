package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort     string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string
}

var App Config

func Load() {
	App = Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://oauth_user:oauth_pass@localhost:5433/oauth?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "supersecretkey"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
	}

	log.Println("✅ Configuration loaded")
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}
