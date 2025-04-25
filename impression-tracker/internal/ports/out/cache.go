package out

import "context"

// CacheRepository gère le compteur en cache (Dragonfly)
type CacheRepository interface {
	Increment(ctx context.Context, adID string) (int64, error)
	Get(ctx context.Context, adID string) (int64, error)
	Reset(ctx context.Context, adID string) (int64, error)
	GetAllKeys(ctx context.Context) ([]string, error)
}
