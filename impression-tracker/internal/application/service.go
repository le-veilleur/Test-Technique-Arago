package application

import (
	"context"
	"log"
	"sync"
	"time"

	"impression-tracker/internal/ports/out"
)

// Service implémente la logique métier du suivi d'impressions.
// Il gère la synchronisation périodique entre le cache (Dragonfly) et le stockage persistant (MongoDB).
type Service struct {
	cacheRepo  out.CacheRepository   // Repository pour le cache (Dragonfly)
	storeRepo  out.MetricsRepository // Repository pour le stockage persistant (MongoDB)
	syncTicker *time.Ticker          // Timer pour la synchronisation périodique
	stopChan   chan struct{}         // Canal pour arrêter la synchronisation
	wg         sync.WaitGroup        // WaitGroup pour gérer la goroutine de synchronisation
}

// NewService crée une nouvelle instance de Service.
// Elle initialise les repositories et configure la synchronisation périodique.
func NewService(cacheRepo out.CacheRepository, storeRepo out.MetricsRepository, syncInterval time.Duration) *Service {
	return &Service{
		cacheRepo:  cacheRepo,
		storeRepo:  storeRepo,
		syncTicker: time.NewTicker(syncInterval),
		stopChan:   make(chan struct{}),
	}
}

// Start démarre la goroutine de synchronisation périodique.
func (s *Service) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.syncTicker.C:
				s.sync()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop arrête la goroutine de synchronisation et attend sa terminaison.
func (s *Service) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	s.syncTicker.Stop()
}

// TrackImpression incrémente le compteur d'impressions pour une publicité donnée.
func (s *Service) TrackImpression(ctx context.Context, adID string) error {
	_, err := s.cacheRepo.Increment(ctx, adID)
	return err
}

// GetImpressionCount récupère le nombre actuel d'impressions pour une publicité donnée.
func (s *Service) GetImpressionCount(ctx context.Context, adID string) (int64, error) {
	return s.cacheRepo.Get(ctx, adID)
}

// Track incrémente le compteur d'impressions pour une publicité donnée.
// Implémente l'interface in.ImpressionService.
func (s *Service) Track(ctx context.Context, adID string) error {
	return s.TrackImpression(ctx, adID)
}

// GetCount récupère le nombre actuel d'impressions pour une publicité donnée.
// Implémente l'interface in.ImpressionService.
func (s *Service) GetCount(ctx context.Context, adID string) (int64, error) {
	return s.GetImpressionCount(ctx, adID)
}

// sync synchronise les compteurs d'impressions entre le cache et le stockage persistant.
// Pour chaque publicité :
// 1. Récupère et réinitialise le compteur dans le cache
// 2. Si le compteur est > 0, persiste le delta dans MongoDB
func (s *Service) sync() {
	ctx := context.Background()

	// Get all ad IDs from cache
	adIDs, err := s.cacheRepo.GetAllKeys(ctx)
	if err != nil {
		log.Printf("Error getting keys from cache: %v", err)
		return
	}

	// Process each ad ID
	for _, adID := range adIDs {
		// Get and reset the count in cache
		count, err := s.cacheRepo.Reset(ctx, adID)
		if err != nil {
			log.Printf("Error resetting count for ad %s: %v", adID, err)
			continue
		}

		// If there were impressions, persist the delta
		if count > 0 {
			if err := s.storeRepo.PersistDelta(ctx, adID, count); err != nil {
				log.Printf("Error persisting delta for ad %s: %v", adID, err)
				continue
			}
			log.Printf("Synced %d impressions for ad %s", count, adID)
		}
	}
}
