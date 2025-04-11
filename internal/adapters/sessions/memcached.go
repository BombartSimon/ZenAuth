package adapters

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemcachedLimiter implémente le rate limiting avec Memcached
type MemcachedLimiter struct {
	client *memcache.Client
	config LimiterConfig
}

func NewMemcachedLimiter(servers []string, config LimiterConfig) (*MemcachedLimiter, error) {
	limiter := &MemcachedLimiter{
		client: memcache.New(servers...),
		config: config,
	}

	// Test connection
	if err := limiter.client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to memcached: %w", err)
	}

	// Démarrer la tâche de nettoyage en arrière-plan
	limiter.StartCleanupTask()

	return limiter, nil
}

// RecordFailedAttempt incrémente le compteur pour les tentatives de connexion échouées
func (m *MemcachedLimiter) RecordFailedAttempt(identifier string) (int, error) {
	attemptsKey := fmt.Sprintf("failed_attempts:%s", identifier)

	// Récupérer le nombre actuel de tentatives
	item, err := m.client.Get(attemptsKey)
	var attempts int

	// Si la clé n'existe pas, commencer à 1
	if err == memcache.ErrCacheMiss {
		attempts = 1
	} else if err != nil {
		return 0, fmt.Errorf("failed to get attempts count: %w", err)
	} else {
		// Parser la valeur existante
		attempts, err = strconv.Atoi(string(item.Value))
		if err != nil {
			return 0, fmt.Errorf("invalid attempts count: %w", err)
		}
		attempts++
	}

	// Stocker la valeur incrémentée
	err = m.client.Set(&memcache.Item{
		Key:        attemptsKey,
		Value:      []byte(strconv.Itoa(attempts)),
		Expiration: int32(m.config.CounterExpiration.Seconds()),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to store attempts count: %w", err)
	}

	// Si le nombre maximum de tentatives est atteint, bloquer l'utilisateur
	if attempts >= m.config.MaxAttempts {
		blockKey := fmt.Sprintf("blocked:%s", identifier)
		err = m.client.Set(&memcache.Item{
			Key:        blockKey,
			Value:      []byte("blocked"),
			Expiration: int32(m.config.BlockDuration.Seconds()),
		})
		if err != nil {
			return attempts, fmt.Errorf("failed to block user: %w", err)
		}

		if err := m.addToBlockedList(identifier); err != nil {
			log.Printf("Warning: failed to update blocked list: %v", err)
		}
	}

	return attempts, nil
}

func (m *MemcachedLimiter) addToBlockedList(identifier string) error {
	const blockedListKey = "blocked_identifiers_list"

	// Utiliser un verrou pour éviter les problèmes de concurrence
	lockKey := "lock_blocked_list"
	err := m.client.Add(&memcache.Item{
		Key:        lockKey,
		Value:      []byte("locked"),
		Expiration: 5, // 5 secondes max pour le verrou
	})

	// Si le verrou existe déjà, on attend un peu et on réessaie
	if err == memcache.ErrNotStored {
		time.Sleep(100 * time.Millisecond)
		return m.addToBlockedList(identifier)
	} else if err != nil && err != memcache.ErrNotStored {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Libérer le verrou à la fin
	defer m.client.Delete(lockKey)

	// Récupérer la liste actuelle
	var blockedList []string
	item, err := m.client.Get(blockedListKey)
	if err == nil {
		// La liste existe, on la désérialise
		if err := json.Unmarshal(item.Value, &blockedList); err != nil {
			return fmt.Errorf("failed to unmarshal blocked list: %w", err)
		}
	} else if err != memcache.ErrCacheMiss {
		return fmt.Errorf("failed to get blocked list: %w", err)
	}

	// Vérifier si l'identifiant est déjà dans la liste
	for _, id := range blockedList {
		if id == identifier {
			return nil // Déjà présent
		}
	}

	// Ajouter l'identifiant à la liste
	blockedList = append(blockedList, identifier)

	// Sérialiser et stocker la liste mise à jour
	jsonData, err := json.Marshal(blockedList)
	if err != nil {
		return fmt.Errorf("failed to marshal blocked list: %w", err)
	}

	return m.client.Set(&memcache.Item{
		Key:        blockedListKey,
		Value:      jsonData,
		Expiration: int32((24 * time.Hour).Seconds()), // Expire après 24h
	})
}

// IsBlocked vérifie si un identifiant est actuellement bloqué
func (m *MemcachedLimiter) IsBlocked(identifier string) (bool, error) {
	blockKey := fmt.Sprintf("blocked:%s", identifier)

	_, err := m.client.Get(blockKey)
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check blocked status: %w", err)
	}

	return true, nil
}

// Reset réinitialise les compteurs pour un identifiant
func (m *MemcachedLimiter) Reset(identifier string) error {
	attemptsKey := fmt.Sprintf("failed_attempts:%s", identifier)
	blockKey := fmt.Sprintf("blocked:%s", identifier)

	// Supprimer les deux clés
	if err := m.client.Delete(attemptsKey); err != nil && err != memcache.ErrCacheMiss {
		return fmt.Errorf("failed to reset attempts: %w", err)
	}

	// Vérifier d'abord si l'utilisateur était bloqué
	wasBlocked := false
	_, err := m.client.Get(blockKey)
	if err == nil {
		wasBlocked = true
	}

	if err := m.client.Delete(blockKey); err != nil && err != memcache.ErrCacheMiss {
		return fmt.Errorf("failed to unblock: %w", err)
	}

	// Si l'utilisateur était bloqué, le retirer de la liste
	if wasBlocked {
		if err := m.removeFromBlockedList(identifier); err != nil {
			log.Printf("Warning: failed to update blocked list: %v", err)
		}
	}

	return nil
}

func (m *MemcachedLimiter) removeFromBlockedList(identifier string) error {
	const blockedListKey = "blocked_identifiers_list"

	// Utiliser un verrou comme dans addToBlockedList
	lockKey := "lock_blocked_list"
	err := m.client.Add(&memcache.Item{
		Key:        lockKey,
		Value:      []byte("locked"),
		Expiration: 5,
	})

	if err == memcache.ErrNotStored {
		time.Sleep(100 * time.Millisecond)
		return m.removeFromBlockedList(identifier)
	} else if err != nil && err != memcache.ErrNotStored {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	defer m.client.Delete(lockKey)

	// Récupérer la liste actuelle
	item, err := m.client.Get(blockedListKey)
	if err == memcache.ErrCacheMiss {
		return nil // Rien à faire
	}
	if err != nil {
		return fmt.Errorf("failed to get blocked list: %w", err)
	}

	var blockedList []string
	if err := json.Unmarshal(item.Value, &blockedList); err != nil {
		return fmt.Errorf("failed to unmarshal blocked list: %w", err)
	}

	// Retirer l'identifiant de la liste
	newList := make([]string, 0, len(blockedList))
	for _, id := range blockedList {
		if id != identifier {
			newList = append(newList, id)
		}
	}

	// Si la liste est vide, supprimer la clé
	if len(newList) == 0 {
		return m.client.Delete(blockedListKey)
	}

	// Sérialiser et stocker la liste mise à jour
	jsonData, err := json.Marshal(newList)
	if err != nil {
		return fmt.Errorf("failed to marshal blocked list: %w", err)
	}

	return m.client.Set(&memcache.Item{
		Key:        blockedListKey,
		Value:      jsonData,
		Expiration: int32((24 * time.Hour).Seconds()),
	})
}

// GetMaxAttempts retourne le nombre maximum de tentatives configuré
func (m *MemcachedLimiter) GetMaxAttempts() int {
	return m.config.MaxAttempts
}

// GetBlockDuration retourne la durée de blocage configurée
func (m *MemcachedLimiter) GetBlockDuration() time.Duration {
	return m.config.BlockDuration
}

func (m *MemcachedLimiter) GetBlockedIdentifiers() ([]string, error) {
	const blockedListKey = "blocked_identifiers_list"

	// Récupérer la liste des identifiants bloqués
	item, err := m.client.Get(blockedListKey)
	if err == memcache.ErrCacheMiss {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get blocked identifiers: %w", err)
	}

	var blockedList []string
	if err := json.Unmarshal(item.Value, &blockedList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blocked list: %w", err)
	}

	// Vérifier que chaque identifiant est toujours bloqué
	// (pour gérer les cas où le blocage a expiré)
	result := make([]string, 0, len(blockedList))
	for _, id := range blockedList {
		blockKey := fmt.Sprintf("blocked:%s", id)
		_, err := m.client.Get(blockKey)
		if err == nil {
			// L'utilisateur est toujours bloqué
			result = append(result, id)
		} else if err != memcache.ErrCacheMiss {
			log.Printf("Error checking block status for %s: %v", id, err)
		}
	}

	// Si la liste a changé, mettre à jour la liste stockée
	if len(result) != len(blockedList) {
		jsonData, err := json.Marshal(result)
		if err != nil {
			log.Printf("Error marshaling updated blocked list: %v", err)
		} else {
			m.client.Set(&memcache.Item{
				Key:        blockedListKey,
				Value:      jsonData,
				Expiration: int32((24 * time.Hour).Seconds()),
			})
		}
	}

	return result, nil
}

// GetRemainingBlockTime renvoie le temps restant avant déblocage
func (m *MemcachedLimiter) GetRemainingBlockTime(identifier string) (string, error) {
	// Cette fonctionnalité n'est pas directement supportée par Memcached
	// Nous retournons une estimation basée sur la durée configurée
	return fmt.Sprintf("~%d minutes", int(m.config.BlockDuration.Minutes())), nil
}

func (m *MemcachedLimiter) StartCleanupTask() {
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			// Cette opération force une vérification et nettoyage de la liste
			_, err := m.GetBlockedIdentifiers()
			if err != nil {
				log.Printf("Error during blocked list cleanup: %v", err)
			}
		}
	}()
}

// RecordUserIP enregistre l'association entre un utilisateur et une adresse IP
func (m *MemcachedLimiter) RecordUserIP(username, ip string) error {
	// Enregistrer l'IP pour cet utilisateur
	userKey := fmt.Sprintf("user_ips:%s", username)
	return m.addToSet(userKey, ip, 72*time.Hour) // Conserver 72h
}

// GetIPsForUser renvoie toutes les IPs associées à un utilisateur
func (m *MemcachedLimiter) GetIPsForUser(username string) ([]string, error) {
	userKey := fmt.Sprintf("user_ips:%s", username)
	return m.getSetMembers(userKey)
}

// GetUsersForIP renvoie tous les utilisateurs associés à une IP
func (m *MemcachedLimiter) GetUsersForIP(ip string) ([]string, error) {
	ipKey := fmt.Sprintf("ip_users:%s", ip)
	return m.getSetMembers(ipKey)
}

// addToSet est une fonction utilitaire pour ajouter un élément à un ensemble stocké dans Memcached
func (m *MemcachedLimiter) addToSet(key, value string, expiration time.Duration) error {
	// Utiliser un verrou pour éviter les problèmes de concurrence
	lockKey := fmt.Sprintf("lock_%s", key)
	err := m.client.Add(&memcache.Item{
		Key:        lockKey,
		Value:      []byte("locked"),
		Expiration: 5, // 5 secondes max pour le verrou
	})

	if err == memcache.ErrNotStored {
		time.Sleep(100 * time.Millisecond)
		return m.addToSet(key, value, expiration)
	} else if err != nil && err != memcache.ErrNotStored {
		return fmt.Errorf("failed to acquire lock for %s: %w", key, err)
	}

	defer m.client.Delete(lockKey)

	// Récupérer l'ensemble actuel
	var set []string
	item, err := m.client.Get(key)
	if err == nil {
		if err := json.Unmarshal(item.Value, &set); err != nil {
			return fmt.Errorf("failed to unmarshal set %s: %w", key, err)
		}
	} else if err != memcache.ErrCacheMiss {
		return fmt.Errorf("failed to get set %s: %w", key, err)
	}

	// Vérifier si la valeur existe déjà
	for _, existingValue := range set {
		if existingValue == value {
			return nil // Valeur déjà présente
		}
	}

	// Ajouter la valeur et enregistrer
	set = append(set, value)
	jsonData, err := json.Marshal(set)
	if err != nil {
		return fmt.Errorf("failed to marshal set %s: %w", key, err)
	}

	// Enregistrer avec expiration
	err = m.client.Set(&memcache.Item{
		Key:        key,
		Value:      jsonData,
		Expiration: int32(expiration.Seconds()),
	})
	if err != nil {
		return fmt.Errorf("failed to store set %s: %w", key, err)
	}

	// Si c'est une association utilisateur-IP, créer aussi l'association inverse
	if strings.HasPrefix(key, "user_ips:") {
		username := strings.TrimPrefix(key, "user_ips:")
		ipKey := fmt.Sprintf("ip_users:%s", value)
		if err := m.addToSet(ipKey, username, expiration); err != nil {
			log.Printf("Warning: failed to set reverse mapping for %s->%s: %v", value, username, err)
		}
	}

	return nil
}

// getSetMembers récupère les membres d'un ensemble stocké dans Memcached
func (m *MemcachedLimiter) getSetMembers(key string) ([]string, error) {
	item, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get set %s: %w", key, err)
	}

	var set []string
	if err := json.Unmarshal(item.Value, &set); err != nil {
		return nil, fmt.Errorf("failed to unmarshal set %s: %w", key, err)
	}

	return set, nil
}
