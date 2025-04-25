package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"adserver/internal/domain"
	"adserver/internal/ports/in"
	"adserver/internal/ports/out"

	"github.com/google/uuid"
)

// AdServiceImpl implémente l'interface AdService
// Cette implémentation gère la logique métier des annonces
type AdServiceImpl struct {
	repo out.AdRepository
}

// NewAdService crée une nouvelle instance du service d'annonces
func NewAdService(repo out.AdRepository) in.AdService {
	return &AdServiceImpl{repo: repo}
}

// CreateAd crée une nouvelle annonce
func (s *AdServiceImpl) CreateAd(ctx context.Context, ad *domain.Pub) (*domain.Pub, error) {
	start := time.Now()
	log.Printf("[AdService CreateAd] start: Title=%s", ad.Title)

	// Génération d'un ID unique
	ad.ID = uuid.New()

	// Construction de l'URL de tracking
	ad.URL = fmt.Sprintf("https://%s/ads/%s", "localhost:8080", ad.ID.String())

	// Si pas de date d'expiration, on met une date par défaut (24h)
	if ad.ExpiresAt.IsZero() {
		ad.ExpiresAt = time.Now().Add(24 * time.Hour)
	}

	// Initialisation du compteur d'impressions
	ad.Impressions = 0

	// Validation de la date d'expiration
	if !ad.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("expiration date must be in the future")
	}

	// Création dans le repository
	_, err := s.repo.Create(ctx, ad)
	if err != nil {
		log.Printf("[AdService CreateAd] error: %v", err)
		return nil, err
	}

	log.Printf("[AdService CreateAd] completed in %v id=%s", time.Since(start), ad.ID)
	return ad, nil
}

// GetAd récupère une annonce par son ID
func (s *AdServiceImpl) GetAd(ctx context.Context, id string) (*domain.Pub, error) {
	start := time.Now()
	log.Printf("[AdService GetAd] start: id=%s", id)

	// Validation de l'ID
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %v", err)
	}

	// Récupération depuis le repository
	ad, err := s.repo.GetByID(ctx, uuid)
	if err != nil {
		log.Printf("[AdService GetAd] error: %v", err)
		return nil, err
	}

	log.Printf("[AdService GetAd] completed in %v id=%s", time.Since(start), id)
	return ad, nil
}

// ServeAd sert une annonce et incrémente son compteur d'impressions
func (s *AdServiceImpl) ServeAd(ctx context.Context, id uuid.UUID) (string, int64, error) {
	start := time.Now()
	log.Printf("[AdService ServeAd] start: id=%s", id)

	// Récupération de l'annonce
	ad, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[AdService ServeAd] error getting ad: %v", err)
		return "", 0, err
	}

	// Vérification de l'expiration
	if !ad.ExpiresAt.After(time.Now()) {
		return "", 0, fmt.Errorf("ad has expired")
	}

	// Incrémentation du compteur d'impressions
	impressions, err := s.repo.IncrementImpressions(ctx, id)
	if err != nil {
		log.Printf("[AdService ServeAd] error incrementing impressions: %v", err)
		return "", 0, err
	}

	log.Printf("[AdService ServeAd] completed in %v id=%s impressions=%d", time.Since(start), id, impressions)
	return ad.URL, impressions, nil
}

// GetAdImpressions récupère le nombre d'impressions d'une annonce
func (s *AdServiceImpl) GetAdImpressions(ctx context.Context, id uuid.UUID) (int64, error) {
	start := time.Now()
	log.Printf("[AdService GetAdImpressions] start: id=%s", id)

	// Récupération du compteur d'impressions
	impressions, err := s.repo.GetImpressions(ctx, id)
	if err != nil {
		log.Printf("[AdService GetAdImpressions] error: %v", err)
		return 0, err
	}

	log.Printf("[AdService GetAdImpressions] completed in %v id=%s impressions=%d", time.Since(start), id, impressions)
	return impressions, nil
}

// CleanupExpired supprime les annonces expirées
func (s *AdServiceImpl) CleanupExpired(ctx context.Context) error {
	_, err := s.DeleteExpired(ctx)
	return err
}

// IncrementImpressions incrémente le compteur d'impressions d'une annonce
func (s *AdServiceImpl) IncrementImpressions(ctx context.Context, id string) (int64, error) {
	start := time.Now()
	log.Printf("[AdService IncrementImpressions] start: id=%s", id)

	// Validation de l'ID
	uuid, err := uuid.Parse(id)
	if err != nil {
		return 0, fmt.Errorf("invalid id format: %v", err)
	}

	// Incrémentation du compteur
	impressions, err := s.repo.IncrementImpressions(ctx, uuid)
	if err != nil {
		log.Printf("[AdService IncrementImpressions] error: %v", err)
		return 0, err
	}

	log.Printf("[AdService IncrementImpressions] completed in %v id=%s impressions=%d", time.Since(start), id, impressions)
	return impressions, nil
}

// DeleteExpired supprime les annonces expirées
func (s *AdServiceImpl) DeleteExpired(ctx context.Context) (int64, error) {
	start := time.Now()
	log.Printf("[AdService DeleteExpired] start")

	// Suppression des annonces expirées
	count, err := s.repo.DeleteExpired(ctx)
	if err != nil {
		log.Printf("[AdService DeleteExpired] error: %v", err)
		return 0, err
	}

	log.Printf("[AdService DeleteExpired] completed in %v deleted=%d", time.Since(start), count)
	return count, nil
}
