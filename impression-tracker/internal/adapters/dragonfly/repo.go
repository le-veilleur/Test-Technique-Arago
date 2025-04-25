package dragonfly

import (
	"context"
	"fmt"
	"strings"
	"time"

	"impression-tracker/internal/ports/out"

	"github.com/redis/go-redis/v9"
)

// DragonflyRepository implémente l'interface CacheRepository pour stocker les compteurs d'impressions
// en utilisant Dragonfly (compatible Redis) comme cache.
type DragonflyRepository struct {
	client *redis.Client
}

// NewDragonflyRepository crée une nouvelle instance de DragonflyRepository.
// Elle établit une connexion avec le serveur Dragonfly et vérifie que la connexion fonctionne.
func NewDragonflyRepository(addr string) (*DragonflyRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Dragonfly: %w", err)
	}

	return &DragonflyRepository{
		client: client,
	}, nil
}

// Increment incrémente le compteur d'impressions pour une publicité donnée.
// La clé est formatée comme "impression:{adID}" pour éviter les collisions.
func (r *DragonflyRepository) Increment(ctx context.Context, adID string) (int64, error) {
	key := fmt.Sprintf("impression:%s", adID)
	return r.client.Incr(ctx, key).Result()
}

// Get récupère le nombre actuel d'impressions pour une publicité donnée.
// Retourne 0 si la clé n'existe pas.
func (r *DragonflyRepository) Get(ctx context.Context, adID string) (int64, error) {
	key := fmt.Sprintf("impression:%s", adID)
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// Reset récupère et réinitialise le compteur d'impressions pour une publicité donnée.
// Utilise une transaction pipeline pour garantir l'atomicité de l'opération.
func (r *DragonflyRepository) Reset(ctx context.Context, adID string) (int64, error) {
	key := fmt.Sprintf("impression:%s", adID)
	pipe := r.client.Pipeline()
	getCmd := pipe.Get(ctx, key)
	pipe.Del(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, err
	}

	count, err := getCmd.Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// GetAllKeys récupère toutes les clés d'impressions stockées dans Dragonfly.
// Utilise SCAN pour itérer sur toutes les clés de manière efficace, même avec un grand nombre de clés.
func (r *DragonflyRepository) GetAllKeys(ctx context.Context) ([]string, error) {
	// Use SCAN to get all keys matching the pattern
	var cursor uint64
	var keys []string
	for {
		var err error
		var partialKeys []string
		partialKeys, cursor, err = r.client.Scan(ctx, cursor, "impression:*", 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, partialKeys...)
		if cursor == 0 {
			break
		}
	}

	// Extract ad IDs from keys
	adIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			adIDs = append(adIDs, parts[1])
		}
	}

	return adIDs, nil
}

// Close ferme la connexion avec le serveur Dragonfly.
func (r *DragonflyRepository) Close() error {
	return r.client.Close()
}

// Ensure DragonflyRepository implements the CacheRepository interface
var _ out.CacheRepository = (*DragonflyRepository)(nil)
 