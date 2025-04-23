package out

import (
	"adserver/internal/domain"
	"context"

	"github.com/google/uuid"
)

// AdRepository définit les opérations de stockage pour les publicités
// et encapsule les détails de persistence (MongoDB, Dragonfly, etc.).
// Toutes les méthodes opèrent sur l'agrégat métier domain.Pub.

type AdRepository interface {
	// Create insère une nouvelle publicité et retourne l'ID généré.
	Create(ctx context.Context, ad *domain.Pub) (string, error)

	// GetByID récupère une publicité par son ID.
	// Retourne nil,nil si aucune publicité trouvée.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Pub, error)

	// Exists vérifie l'existence d'une publicité sans charger tout l'objet.
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// IncrementImpressions incrémente atomiquement le compteur d'impressions
	// et retourne la nouvelle valeur du compteur.
	IncrementImpressions(ctx context.Context, id uuid.UUID) (newCount int64, err error)

	// ResetImpressions réinitialise le compteur en cache (pour batch sync)
	// et retourne l'ancienne valeur avant reset.
	ResetImpressions(ctx context.Context, id uuid.UUID) (oldCount int64, err error)

	// DeleteExpired supprime toutes les publicités expirées.
	// Retourne le nombre de documents supprimés.
	DeleteExpired(ctx context.Context) (deletedCount int64, err error)

	// List permet de lister les publicités selon un filtre et pagination.
	List(ctx context.Context, filter map[string]interface{}, offset, limit int64) ([]*domain.Pub, error)
}
