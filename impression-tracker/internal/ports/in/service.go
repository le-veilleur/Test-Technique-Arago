package in

import "context"

// ImpressionService définit les cas d'usage du suivi d'impressions.
// C'est le port d'entrée (primary port) de l'application.
type ImpressionService interface {
	// Track enregistre une nouvelle impression pour une publicité donnée
	Track(ctx context.Context, adID string) error

	// GetCount récupère le nombre d'impressions pour une publicité donnée
	GetCount(ctx context.Context, adID string) (int64, error)
}
