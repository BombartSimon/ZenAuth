package adapters

import "time"

// Limiter définit l'interface pour le rate limiting
type Limiter interface {
	// RecordFailedAttempt enregistre une tentative échouée et retourne le nombre actuel de tentatives
	RecordFailedAttempt(identifier string) (int, error)

	// IsBlocked vérifie si un identifiant est actuellement bloqué
	IsBlocked(identifier string) (bool, error)

	// Reset réinitialise les compteurs pour un identifiant
	Reset(identifier string) error

	// GetMaxAttempts retourne le nombre maximum de tentatives configuré
	GetMaxAttempts() int

	// GetBlockDuration retourne la durée de blocage configurée
	GetBlockDuration() time.Duration

	GetBlockedIdentifiers() ([]string, error)
	GetRemainingBlockTime(identifier string) (string, error)

	RecordUserIP(username, ip string) error
	GetIPsForUser(username string) ([]string, error)
	GetUsersForIP(ip string) ([]string, error)
}
