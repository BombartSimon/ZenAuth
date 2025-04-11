package adapters

import "time"

// LimiterConfig représente la configuration commune pour les limiters
type LimiterConfig struct {
	// Nombre maximum de tentatives échouées avant blocage
	MaxAttempts int

	// Durée du blocage après dépassement du nombre maximum de tentatives
	BlockDuration time.Duration

	// Durée d'expiration des compteurs de tentatives
	CounterExpiration time.Duration
}
