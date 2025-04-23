package handler

import (
	"context"
	"log"
	"time"

	"adserver/generated/ad_service"
	"adserver/internal/domain"
	"adserver/internal/ports/in"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdHandler implémente le service gRPC AdService
type AdHandler struct {
	adService in.AdService
	ad_service.UnimplementedAdServiceServer
}

// NewAdHandler crée une nouvelle instance du handler
func NewAdHandler(adService in.AdService) *AdHandler {
	return &AdHandler{adService: adService}
}

// CreateAd implémente la création d'une publicité
func (h *AdHandler) CreateAd(ctx context.Context, req *ad_service.CreateAdRequest) (*ad_service.AdResponse, error) {
	start := time.Now()
	log.Printf("[CreateAd] start: title=%q url=%q expiresAt=%v", req.Title, req.Url, req.ExpiresAt)

	if req.Title == "" || req.Url == "" {
		log.Printf("[CreateAd] invalid argument: title or url empty")
		return nil, status.Error(codes.InvalidArgument, "title and url are required")
	}

	var expiresAt time.Time
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt.AsTime()
	} else {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	ad := &domain.Pub{
		Title:       req.Title,
		Description: &req.Description,
		URL:         req.Url,
		ExpiresAt:   expiresAt,
	}

	id, err := h.adService.CreateAd(ctx, ad)
	if err != nil {
		log.Printf("[CreateAd] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ad_service.AdResponse{
		Id:          id,
		Title:       ad.Title,
		Description: *ad.Description,
		Url:         ad.URL,
		ExpiresAt:   timestamppb.New(ad.ExpiresAt),
	}
	log.Printf("[CreateAd] completed in %v id=%s", time.Since(start), id)
	return resp, nil
}

// GetAd implémente la récupération d'une publicité
func (h *AdHandler) GetAd(ctx context.Context, req *ad_service.GetAdRequest) (*ad_service.AdResponse, error) {
	start := time.Now()
	log.Printf("[GetAd] start: id=%q", req.Id)

	if req.Id == "" {
		log.Printf("[GetAd] invalid argument: id empty")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ad, err := h.adService.GetAd(ctx, req.Id)
	if err != nil {
		log.Printf("[GetAd] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &ad_service.AdResponse{
		Id:        ad.ID.String(),
		Title:     ad.Title,
		Url:       ad.URL,
		ExpiresAt: timestamppb.New(ad.ExpiresAt),
	}
	if ad.Description != nil {
		response.Description = *ad.Description
	}
	log.Printf("[GetAd] completed in %v id=%s", time.Since(start), req.Id)
	return response, nil
}

// ServeAd implémente la diffusion d'une publicité
func (h *AdHandler) ServeAd(ctx context.Context, req *ad_service.ServeAdRequest) (*ad_service.ServeAdResponse, error) {
	start := time.Now()
	log.Printf("[ServeAd] start: id=%q", req.Id)

	if req.Id == "" {
		log.Printf("[ServeAd] invalid argument: id empty")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		log.Printf("[ServeAd] invalid id format: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid id format")
	}

	url, impressions, err := h.adService.ServeAd(ctx, id)
	if err != nil {
		log.Printf("[ServeAd] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ad_service.ServeAdResponse{Url: url, Impressions: impressions}
	log.Printf("[ServeAd] completed in %v id=%s impressions=%d", time.Since(start), req.Id, impressions)
	return resp, nil
}

// GetImpressionCount implémente la récupération du nombre d'impressions d'une publicité
func (h *AdHandler) GetImpressionCount(ctx context.Context, req *ad_service.GetImpressionCountRequest) (*ad_service.GetImpressionCountResponse, error) {
	start := time.Now()
	log.Printf("[GetImpressionCount] start: adId=%q", req.AdId)

	if req.AdId == "" {
		log.Printf("[GetImpressionCount] invalid argument: adId empty")
		return nil, status.Error(codes.InvalidArgument, "ad_id is required")
	}

	id, err := uuid.Parse(req.AdId)
	if err != nil {
		log.Printf("[GetImpressionCount] invalid ad_id format: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid ad_id format")
	}

	impr, err := h.adService.GetAdImpressions(ctx, id)
	if err != nil {
		log.Printf("[GetImpressionCount] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ad_service.GetImpressionCountResponse{Impressions: impr}
	log.Printf("[GetImpressionCount] completed in %v adId=%s impressions=%d", time.Since(start), req.AdId, impr)
	return resp, nil
}

// IncrementImpressions implémente l'incrémentation du compteur d'impressions
func (h *AdHandler) IncrementImpressions(ctx context.Context, req *ad_service.IncrementImpressionsRequest) (*ad_service.IncrementImpressionsResponse, error) {
	start := time.Now()
	log.Printf("[IncrementImpressions] start: adId=%q", req.AdId)

	if req.AdId == "" {
		log.Printf("[IncrementImpressions] invalid argument: adId empty")
		return nil, status.Error(codes.InvalidArgument, "ad_id is required")
	}

	impr, err := h.adService.IncrementImpressions(ctx, req.AdId)
	if err != nil {
		log.Printf("[IncrementImpressions] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ad_service.IncrementImpressionsResponse{Impressions: impr}
	log.Printf("[IncrementImpressions] completed in %v adId=%s impressions=%d", time.Since(start), req.AdId, impr)
	return resp, nil
}

// ResetImpressions implémente la réinitialisation du compteur d'impressions
func (h *AdHandler) ResetImpressions(ctx context.Context, req *ad_service.ResetImpressionsRequest) (*ad_service.ResetImpressionsResponse, error) {
	start := time.Now()
	log.Printf("[ResetImpressions] start: adId=%q", req.AdId)

	if req.AdId == "" {
		log.Printf("[ResetImpressions] invalid argument: adId empty")
		return nil, status.Error(codes.InvalidArgument, "ad_id is required")
	}

	impr, err := h.adService.ResetImpressions(ctx, req.AdId)
	if err != nil {
		log.Printf("[ResetImpressions] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ad_service.ResetImpressionsResponse{Impressions: impr}
	log.Printf("[ResetImpressions] completed in %v adId=%s impressions=%d", time.Since(start), req.AdId, impr)
	return resp, nil
}

// DeleteExpired implémente la suppression des annonces expirées
func (h *AdHandler) DeleteExpired(ctx context.Context, req *ad_service.DeleteExpiredRequest) (*ad_service.DeleteExpiredResponse, error) {
	start := time.Now()
	log.Printf("[DeleteExpired] start")

	count, err := h.adService.DeleteExpired(ctx)
	if err != nil {
		log.Printf("[DeleteExpired] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &ad_service.DeleteExpiredResponse{DeletedCount: count}
	log.Printf("[DeleteExpired] completed in %v deletedCount=%d", time.Since(start), count)
	return resp, nil
}

// ListAds implémente la récupération d'une liste paginée d'annonces
func (h *AdHandler) ListAds(ctx context.Context, req *ad_service.ListAdsRequest) (*ad_service.ListAdsResponse, error) {
	start := time.Now()
	log.Printf("[ListAds] start: offset=%d limit=%d filters=%v", req.Offset, req.Limit, req.Filter)

	filter := make(map[string]interface{})
	if req.Filter != nil {
		for k, v := range req.Filter {
			filter[k] = v
		}
	}

	ads, err := h.adService.ListAds(ctx, filter, req.Offset, req.Limit)
	if err != nil {
		log.Printf("[ListAds] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &ad_service.ListAdsResponse{Ads: make([]*ad_service.AdResponse, len(ads))}
	for i, ad := range ads {
		response.Ads[i] = &ad_service.AdResponse{
			Id:        ad.ID.String(),
			Title:     ad.Title,
			Url:       ad.URL,
			ExpiresAt: timestamppb.New(ad.ExpiresAt),
		}
		if ad.Description != nil {
			response.Ads[i].Description = *ad.Description
		}
	}

	log.Printf("[ListAds] completed in %v returned=%d", time.Since(start), len(ads))
	return response, nil
}
