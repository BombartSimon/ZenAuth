package adapters

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client *redis.Client
	config LimiterConfig
}

func NewRedisLimiter(redisURL string, config LimiterConfig) (*RedisLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisLimiter{
		client: client,
		config: config,
	}, nil
}

func (r *RedisLimiter) RecordFailedAttempt(identifier string) (int, error) {
	ctx := context.Background()
	attemptsKey := fmt.Sprintf("failed_attempts:%s", identifier)

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, attemptsKey)
	pipe.Expire(ctx, attemptsKey, r.config.CounterExpiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment attempts count: %w", err)
	}

	attempts := int(incr.Val())

	if attempts >= r.config.MaxAttempts {
		blockKey := fmt.Sprintf("blocked:%s", identifier)
		err = r.client.Set(ctx, blockKey, "blocked", r.config.BlockDuration).Err()
		if err != nil {
			return attempts, fmt.Errorf("failed to block user: %w", err)
		}
	}

	return attempts, nil
}

func (r *RedisLimiter) IsBlocked(identifier string) (bool, error) {
	ctx := context.Background()
	blockKey := fmt.Sprintf("blocked:%s", identifier)

	exists, err := r.client.Exists(ctx, blockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blocked status: %w", err)
	}

	return exists > 0, nil
}

func (r *RedisLimiter) Reset(identifier string) error {
	ctx := context.Background()
	keysToDelete := []string{
		fmt.Sprintf("failed_attempts:%s", identifier),
		fmt.Sprintf("blocked:%s", identifier),
	}

	if strings.HasPrefix(identifier, "user:") {
		username := strings.TrimPrefix(identifier, "user:")
		log.Printf("Searching for IPs associated with user '%s'", username)

		ips, err := r.GetIPsForUser(username)
		if err != nil {
			log.Printf("Error retrieving IPs for user '%s': %v", username, err)
		} else {
			log.Printf("IPs found for user '%s': %v", username, ips)

			for _, ip := range ips {
				ipFailedKey := fmt.Sprintf("failed_attempts:%s", ip)
				ipBlockedKey := fmt.Sprintf("blocked:%s", ip)
				keysToDelete = append(keysToDelete, ipFailedKey, ipBlockedKey)

				log.Printf("Adding keys to delete for IP '%s': %s, %s", ip, ipFailedKey, ipBlockedKey)
			}

			keysToDelete = append(keysToDelete, fmt.Sprintf("user_ips:%s", username))
		}
	} else {
		log.Printf("Searching for users associated with IP '%s'", identifier)

		users, err := r.GetUsersForIP(identifier)
		if err != nil {
			log.Printf("Error retrieving users for IP '%s': %v", identifier, err)
		} else {
			log.Printf("Users found for IP '%s': %v", identifier, users)

			for _, user := range users {
				userFailedKey := fmt.Sprintf("failed_attempts:user:%s", user)
				userBlockedKey := fmt.Sprintf("blocked:user:%s", user)
				keysToDelete = append(keysToDelete, userFailedKey, userBlockedKey)

				log.Printf("Adding keys to delete for user '%s': %s, %s", user, userFailedKey, userBlockedKey)
			}

			keysToDelete = append(keysToDelete, fmt.Sprintf("ip_users:%s", identifier))
		}
	}

	if len(keysToDelete) > 0 {
		log.Printf("Deleting keys: %v", keysToDelete)
		err := r.client.Del(ctx, keysToDelete...).Err()
		if err != nil {
			return fmt.Errorf("error deleting keys: %w", err)
		}
	}

	return nil
}

func (r *RedisLimiter) GetMaxAttempts() int {
	return r.config.MaxAttempts
}

func (r *RedisLimiter) GetBlockDuration() time.Duration {
	return r.config.BlockDuration
}

func (r *RedisLimiter) GetBlockedIdentifiers() ([]string, error) {
	ctx := context.Background()
	keys, err := r.client.Keys(ctx, "blocked:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get blocked identifiers: %w", err)
	}

	identifiers := make([]string, len(keys))
	for i, key := range keys {
		identifiers[i] = key[len("blocked:"):]
	}

	return identifiers, nil
}

func (r *RedisLimiter) GetRemainingBlockTime(identifier string) (string, error) {
	ctx := context.Background()
	blockKey := fmt.Sprintf("blocked:%s", identifier)

	ttl, err := r.client.TTL(ctx, blockKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get remaining block time: %w", err)
	}

	if ttl < 0 {
		return "not blocked", nil
	}

	return ttl.String(), nil
}

func (r *RedisLimiter) RecordUserIP(username, ipAddress string) error {
	ctx := context.Background()

	userIPsKey := fmt.Sprintf("user_ips:%s", username)
	ipUsersKey := fmt.Sprintf("ip_users:%s", ipAddress)

	if err := r.client.SAdd(ctx, userIPsKey, ipAddress).Err(); err != nil {
		return fmt.Errorf("error adding IP for user: %w", err)
	}
	r.client.Expire(ctx, userIPsKey, r.config.CounterExpiration)

	if err := r.client.SAdd(ctx, ipUsersKey, username).Err(); err != nil {
		return fmt.Errorf("error adding user for IP: %w", err)
	}
	r.client.Expire(ctx, ipUsersKey, r.config.CounterExpiration)

	log.Printf("Association recorded between user '%s' and IP '%s'", username, ipAddress)
	return nil
}

func (r *RedisLimiter) GetIPsForUser(username string) ([]string, error) {
	ctx := context.Background()
	userIPsKey := fmt.Sprintf("user_ips:%s", username)

	ips, err := r.client.SMembers(ctx, userIPsKey).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving IPs: %w", err)
	}

	return ips, nil
}

func (r *RedisLimiter) GetUsersForIP(ipAddress string) ([]string, error) {
	ctx := context.Background()
	ipUsersKey := fmt.Sprintf("ip_users:%s", ipAddress)

	users, err := r.client.SMembers(ctx, ipUsersKey).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving users: %w", err)
	}

	return users, nil
}
