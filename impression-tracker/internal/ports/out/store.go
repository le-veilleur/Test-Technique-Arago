package out

import "context"

// MetricsRepository persiste les deltas d'impressions en base
type MetricsRepository interface {
	PersistDelta(ctx context.Context, adID string, delta int64) error
}
