package application

import (
	"context"
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
func (s *AdServiceImpl) CreateAd(ctx context.Context, ad *domain.Pub) (string, error) {
	start := time.Now()
	log.Printf("[AdService CreateAd] start: Title=%s URL=%s ExpiresAt=%v", ad.Title, ad.URL, ad.ExpiresAt)
	if ad.ID == uuid.Nil {
		ad.ID = uuid.New()
	}
	ad.Impressions = 0

	id, err := s.repo.Create(ctx, ad)
	if err != nil {
		log.Printf("[AdService CreateAd] error: %v", err)
		return "", err
	}
	log.Printf("[AdService CreateAd] completed in %v id=%s", time.Since(start), id)
	return id, nil
}

// GetAd récupère une annonce par son ID
func (s *AdServiceImpl) GetAd(ctx context.Context, id string) (*domain.Pub, error) {
	start := time.Now()
	log.Printf("[AdService GetAd] start: id=%s", id)
	u, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[AdService GetAd] invalid id: %v", err)
		return nil, err
	}
	pub, err := s.repo.GetByID(ctx, u)
	if err != nil {
		log.Printf("[AdService GetAd] error: %v", err)
		return nil, err
	}
	log.Printf("[AdService GetAd] completed in %v id=%s", time.Since(start), id)
	return pub, nil
}

// ServeAd sert une annonce et incrémente son compteur d'impressions
func (s *AdServiceImpl) ServeAd(ctx context.Context, id uuid.UUID) (string, int64, error) {
	start := time.Now()
	log.Printf("[AdService ServeAd] start: id=%s", id)
	pub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[AdService ServeAd] GetByID error: %v", err)
		return "", 0, err
	}
	count, err := s.repo.IncrementImpressions(ctx, id)
	if err != nil {
		log.Printf("[AdService ServeAd] IncrementImpressions error: %v", err)
		return "", 0, err
	}
	log.Printf("[AdService ServeAd] completed in %v id=%s impressions=%d", time.Since(start), id, count)
	return pub.URL, count, nil
}

// GetAdImpressions récupère le nombre d'impressions d'une annonce
func (s *AdServiceImpl) GetAdImpressions(ctx context.Context, id uuid.UUID) (int64, error) {
	start := time.Now()
	log.Printf("[AdService GetAdImpressions] start: id=%s", id)
	count, err := s.repo.IncrementImpressions(ctx, id)
	if err != nil {
		log.Printf("[AdService GetAdImpressions] error: %v", err)
		return 0, err
	}
	log.Printf("[AdService GetAdImpressions] completed in %v id=%s impressions=%d", time.Since(start), id, count)
	return count, nil
}

// CleanupExpired supprime les annonces expirées
func (s *AdServiceImpl) CleanupExpired(ctx context.Context) error {
	start := time.Now()
	log.Printf("[AdService CleanupExpired] start")
	count, err := s.repo.DeleteExpired(ctx)
	if err != nil {
		log.Printf("[AdService CleanupExpired] error: %v", err)
		return err
	}
	log.Printf("[AdService CleanupExpired] completed in %v deleted=%d", time.Since(start), count)
	return nil
}

// IncrementImpressions incrémente le compteur d'impressions d'une annonce
func (s *AdServiceImpl) IncrementImpressions(ctx context.Context, id string) (int64, error) {
	start := time.Now()
	log.Printf("[AdService IncrementImpressions] start: id=%s", id)
	u, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[AdService IncrementImpressions] invalid id: %v", err)
		return 0, err
	}
	count, err := s.repo.IncrementImpressions(ctx, u)
	if err != nil {
		log.Printf("[AdService IncrementImpressions] error: %v", err)
		return 0, err
	}
	log.Printf("[AdService IncrementImpressions] completed in %v id=%s impressions=%d", time.Since(start), id, count)
	return count, nil
}

// ResetImpressions réinitialise le compteur d'impressions d'une annonce
func (s *AdServiceImpl) ResetImpressions(ctx context.Context, id string) (int64, error) {
	start := time.Now()
	log.Printf("[AdService ResetImpressions] start: id=%s", id)
	u, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[AdService ResetImpressions] invalid id: %v", err)
		return 0, err
	}
	count, err := s.repo.ResetImpressions(ctx, u)
	if err != nil {
		log.Printf("[AdService ResetImpressions] error: %v", err)
		return 0, err
	}
	log.Printf("[AdService ResetImpressions] completed in %v id=%s oldImpressions=%d", time.Since(start), id, count)
	return count, nil
}

// DeleteExpired supprime les annonces expirées
func (s *AdServiceImpl) DeleteExpired(ctx context.Context) (int64, error) {
	start := time.Now()
	log.Printf("[AdService DeleteExpired] start")
	count, err := s.repo.DeleteExpired(ctx)
	if err != nil {
		log.Printf("[AdService DeleteExpired] error: %v", err)
		return 0, err
	}
	log.Printf("[AdService DeleteExpired] completed in %v deleted=%d", time.Since(start), count)
	return count, nil
}

// ListAds récupère une liste paginée d'annonces avec filtrage optionnel
func (s *AdServiceImpl) ListAds(ctx context.Context, filter map[string]interface{}, offset, limit int64) ([]*domain.Pub, error) {
	start := time.Now()
	log.Printf("[AdService ListAds] start: offset=%d limit=%d filter=%v", offset, limit, filter)
	ads, err := s.repo.List(ctx, filter, offset, limit)
	if err != nil {
		log.Printf("[AdService ListAds] error: %v", err)
		return nil, err
	}
	log.Printf("[AdService ListAds] completed in %v count=%d", time.Since(start), len(ads))
	return ads, nil
}
