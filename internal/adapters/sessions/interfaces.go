package adapters

import "time"

type Limiter interface {
	RecordFailedAttempt(identifier string) (int, error)

	IsBlocked(identifier string) (bool, error)

	Reset(identifier string) error

	GetMaxAttempts() int

	GetBlockDuration() time.Duration

	GetBlockedIdentifiers() ([]string, error)
	GetRemainingBlockTime(identifier string) (string, error)

	RecordUserIP(username, ip string) error
	GetIPsForUser(username string) ([]string, error)
	GetUsersForIP(ip string) ([]string, error)
}
