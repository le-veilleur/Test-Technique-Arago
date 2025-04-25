package in

import (
	"adserver/internal/domain"
	"context"

	"github.com/google/uuid"
)

// AdService définit l'interface pour le service de gestion des annonces
// Cette interface expose les opérations principales pour la gestion des annonces
type AdService interface {
	// CreateAd crée une nouvelle annonce avec les données fournies
	// Retourne l'ID de l'annonce créée ou une erreur
	CreateAd(ctx context.Context, ad *domain.Pub) (*domain.Pub, error)

	// GetAd récupère une annonce par son ID
	// Retourne l'annonce ou une erreur si non trouvée
	GetAd(ctx context.Context, id string) (*domain.Pub, error)

	// ServeAd diffuse la pub, incrémente le compteur, et renvoie :
	// - l'URL à afficher
	// - le nombre d'impressions APRÈS incrément
	ServeAd(ctx context.Context, id uuid.UUID) (string, int64, error)

	// IncrementImpressions incrémente le compteur d'impressions d'une annonce
	// Retourne le nouveau nombre total d'impressions
	IncrementImpressions(ctx context.Context, id string) (int64, error)

	// DeleteExpired supprime toutes les annonces expirées
	// Retourne le nombre d'annonces supprimées
	DeleteExpired(ctx context.Context) (int64, error)

	// GetAdImpressions ne fait QUE lire le compteur, sans toucher au cache ou au track
	GetAdImpressions(ctx context.Context, id uuid.UUID) (int64, error)

	CleanupExpired(ctx context.Context) error
}
