package adapters

import (
	"fmt"
	"log"
	"strings"
	"zenauth/config"
)

var (
	CurrentLimiter Limiter
)

func InitSessions() error {
	if !config.App.RateLimit.Enabled {
		log.Println("Rate limiting is disabled")
		return nil
	}

	limiterConfig := LimiterConfig{
		MaxAttempts:       config.App.RateLimit.MaxAttempts,
		BlockDuration:     config.App.RateLimit.BlockDuration,
		CounterExpiration: config.App.RateLimit.CounterExpiration,
	}

	var limiter Limiter
	var err error

	switch strings.ToLower(config.App.RateLimit.Provider) {
	case "redis":
		limiter, err = NewRedisLimiter(config.App.RateLimit.ConnectionURL, limiterConfig)
	default:
		return fmt.Errorf("unsupported rate limit provider: %s", config.App.RateLimit.Provider)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize rate limiter with %s: %w",
			config.App.RateLimit.Provider, err)
	}

	RegisterLimiter(limiter)
	log.Printf("✅ Rate limiting initialized with %s (max attempts: %d, block duration: %s)",
		config.App.RateLimit.Provider,
		limiter.GetMaxAttempts(),
		limiter.GetBlockDuration())

	return nil
}

// RegisterLimiter initialise et enregistre un limiter
func RegisterLimiter(limiter Limiter) {
	CurrentLimiter = limiter
}

// IsLimiterEnabled vérifie si le rate limiting est activé et configuré
func IsLimiterEnabled() bool {
	return config.App.RateLimit.Enabled && CurrentLimiter != nil
}

// CheckRateLimit vérifie si un identifiant est bloqué
func CheckRateLimit(identifier string) (bool, string, error) {
	if !IsLimiterEnabled() {
		return false, "", nil
	}

	// Vérifier si l'identifiant est bloqué
	blocked, err := CurrentLimiter.IsBlocked(identifier)
	if err != nil {
		return false, "", fmt.Errorf("rate limit check error: %w", err)
	}

	if blocked {
		return true, "Too many login attempts. Please try again later.", nil
	}

	return false, "", nil
}

// RecordFailedLoginAttempt enregistre un échec de connexion et gère le rate limiting
func RecordFailedLoginAttempt(identifier string) (int, error) {
	if !IsLimiterEnabled() {
		return 0, nil
	}

	attempts, err := CurrentLimiter.RecordFailedAttempt(identifier)
	if err != nil {
		return 0, err
	}

	log.Printf("Failed login attempt for '%s': %d/%d attempts",
		identifier, attempts, CurrentLimiter.GetMaxAttempts())

	return attempts, nil
}

// ResetLoginAttempts réinitialise les compteurs après une connexion réussie
func ResetLoginAttempts(identifier string) error {
	if !IsLimiterEnabled() {
		return nil
	}

	if err := CurrentLimiter.Reset(identifier); err != nil {
		return fmt.Errorf("error resetting rate limit for '%s': %w", identifier, err)
	}

	return nil
}

func GetBlockedIdentifiers() ([]string, error) {
	if !IsLimiterEnabled() {
		return []string{}, nil
	}

	return CurrentLimiter.GetBlockedIdentifiers()
}

// GetRemainingBlockTime retourne le temps restant avant déblocage
func GetRemainingBlockTime(identifier string) (string, error) {
	if !IsLimiterEnabled() {
		return "0 minutes", nil
	}

	return CurrentLimiter.GetRemainingBlockTime(identifier)
}
