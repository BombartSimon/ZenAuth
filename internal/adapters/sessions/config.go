package adapters

import "time"

type LimiterConfig struct {
	MaxAttempts int

	BlockDuration time.Duration

	CounterExpiration time.Duration
}
