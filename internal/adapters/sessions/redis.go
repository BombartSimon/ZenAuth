package adapters

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implémente le rate limiting avec Redis
type RedisLimiter struct {
	client *redis.Client
	config LimiterConfig
}

// NewRedisLimiter crée un nouveau rate limiter basé sur Redis
func NewRedisLimiter(redisURL string, config LimiterConfig) (*RedisLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("URL Redis invalide: %w", err)
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	// Tester la connexion
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("échec de connexion à Redis: %w", err)
	}

	return &RedisLimiter{
		client: client,
		config: config,
	}, nil
}

// RecordFailedAttempt incrémente le compteur pour les tentatives de connexion échouées
func (r *RedisLimiter) RecordFailedAttempt(identifier string) (int, error) {
	ctx := context.Background()
	attemptsKey := fmt.Sprintf("failed_attempts:%s", identifier)

	// Incrémenter et récupérer la nouvelle valeur en une seule opération
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, attemptsKey)
	pipe.Expire(ctx, attemptsKey, r.config.CounterExpiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment attempts count: %w", err)
	}

	attempts := int(incr.Val())

	// Si le nombre maximum de tentatives est atteint, bloquer l'utilisateur
	if attempts >= r.config.MaxAttempts {
		blockKey := fmt.Sprintf("blocked:%s", identifier)
		err = r.client.Set(ctx, blockKey, "blocked", r.config.BlockDuration).Err()
		if err != nil {
			return attempts, fmt.Errorf("failed to block user: %w", err)
		}
	}

	return attempts, nil
}

// IsBlocked vérifie si un identifiant est actuellement bloqué
func (r *RedisLimiter) IsBlocked(identifier string) (bool, error) {
	ctx := context.Background()
	blockKey := fmt.Sprintf("blocked:%s", identifier)

	exists, err := r.client.Exists(ctx, blockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blocked status: %w", err)
	}

	return exists > 0, nil
}

// Reset réinitialise les compteurs pour un identifiant
func (r *RedisLimiter) Reset(identifier string) error {
	ctx := context.Background()
	keysToDelete := []string{
		fmt.Sprintf("failed_attempts:%s", identifier),
		fmt.Sprintf("blocked:%s", identifier),
	}

	// S'il s'agit d'un utilisateur, débloquer toutes ses IPs associées
	if strings.HasPrefix(identifier, "user:") {
		username := strings.TrimPrefix(identifier, "user:")
		log.Printf("Recherche des IPs associées à l'utilisateur '%s'", username)

		// Récupérer les IPs associées et les débloquer
		ips, err := r.GetIPsForUser(username)
		if err != nil {
			log.Printf("Erreur lors de la récupération des IPs pour l'utilisateur '%s': %v", username, err)
		} else {
			log.Printf("IPs trouvées pour l'utilisateur '%s': %v", username, ips)

			for _, ip := range ips {
				ipFailedKey := fmt.Sprintf("failed_attempts:%s", ip)
				ipBlockedKey := fmt.Sprintf("blocked:%s", ip)
				keysToDelete = append(keysToDelete, ipFailedKey, ipBlockedKey)

				log.Printf("Ajout des clés à supprimer pour l'IP '%s': %s, %s", ip, ipFailedKey, ipBlockedKey)
			}

			// Supprimer aussi l'association utilisateur-IPs
			keysToDelete = append(keysToDelete, fmt.Sprintf("user_ips:%s", username))
		}
	} else {
		// C'est une IP, débloquer tous ses utilisateurs associés
		log.Printf("Recherche des utilisateurs associés à l'IP '%s'", identifier)

		users, err := r.GetUsersForIP(identifier)
		if err != nil {
			log.Printf("Erreur lors de la récupération des utilisateurs pour l'IP '%s': %v", identifier, err)
		} else {
			log.Printf("Utilisateurs trouvés pour l'IP '%s': %v", identifier, users)

			for _, user := range users {
				userFailedKey := fmt.Sprintf("failed_attempts:user:%s", user)
				userBlockedKey := fmt.Sprintf("blocked:user:%s", user)
				keysToDelete = append(keysToDelete, userFailedKey, userBlockedKey)

				log.Printf("Ajout des clés à supprimer pour l'utilisateur '%s': %s, %s", user, userFailedKey, userBlockedKey)
			}

			// Supprimer aussi l'association IP-utilisateurs
			keysToDelete = append(keysToDelete, fmt.Sprintf("ip_users:%s", identifier))
		}
	}

	// Supprimer toutes les clés en une seule opération
	if len(keysToDelete) > 0 {
		log.Printf("Suppression des clés: %v", keysToDelete)
		err := r.client.Del(ctx, keysToDelete...).Err()
		if err != nil {
			return fmt.Errorf("erreur lors de la suppression des clés: %w", err)
		}
	}

	return nil
}

// GetMaxAttempts retourne le nombre maximum de tentatives configuré
func (r *RedisLimiter) GetMaxAttempts() int {
	return r.config.MaxAttempts
}

// GetBlockDuration retourne la durée de blocage configurée
func (r *RedisLimiter) GetBlockDuration() time.Duration {
	return r.config.BlockDuration
}

// GetBlockedIdentifiers retourne la liste des identifiants bloqués
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

// GetRemainingBlockTime retourne le temps restant de blocage pour un identifiant
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

// RecordUserIP enregistre l'association IP-utilisateur de façon bidirectionnelle
func (r *RedisLimiter) RecordUserIP(username, ipAddress string) error {
	ctx := context.Background()

	// Clés pour les associations bidirectionnelles
	userIPsKey := fmt.Sprintf("user_ips:%s", username)
	ipUsersKey := fmt.Sprintf("ip_users:%s", ipAddress)

	// Ajouter l'IP à l'ensemble des IPs de l'utilisateur
	if err := r.client.SAdd(ctx, userIPsKey, ipAddress).Err(); err != nil {
		return fmt.Errorf("erreur lors de l'ajout de l'IP pour l'utilisateur: %w", err)
	}
	r.client.Expire(ctx, userIPsKey, r.config.CounterExpiration)

	// Ajouter l'utilisateur à l'ensemble des utilisateurs de l'IP
	if err := r.client.SAdd(ctx, ipUsersKey, username).Err(); err != nil {
		return fmt.Errorf("erreur lors de l'ajout de l'utilisateur pour l'IP: %w", err)
	}
	r.client.Expire(ctx, ipUsersKey, r.config.CounterExpiration)

	log.Printf("Association enregistrée entre utilisateur '%s' et IP '%s'", username, ipAddress)
	return nil
}

// GetIPsForUser retourne la liste des adresses IP pour un utilisateur
func (r *RedisLimiter) GetIPsForUser(username string) ([]string, error) {
	ctx := context.Background()
	userIPsKey := fmt.Sprintf("user_ips:%s", username)

	ips, err := r.client.SMembers(ctx, userIPsKey).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des IPs: %w", err)
	}

	return ips, nil
}

// GetUsersForIP retourne la liste des utilisateurs associés à une adresse IP
func (r *RedisLimiter) GetUsersForIP(ipAddress string) ([]string, error) {
	ctx := context.Background()
	ipUsersKey := fmt.Sprintf("ip_users:%s", ipAddress)

	users, err := r.client.SMembers(ctx, ipUsersKey).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des utilisateurs: %w", err)
	}

	return users, nil
}
