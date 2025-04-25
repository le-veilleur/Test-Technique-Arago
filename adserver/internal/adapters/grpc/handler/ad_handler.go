package handler

import (
	"context"
	"log"
	"time"

	"adserver/generated/ad_service"
	"adserver/generated/impression_service"
	"adserver/internal/domain"
	"adserver/internal/ports/in"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdHandler implémente le service gRPC AdService
type AdHandler struct {
	adService        in.AdService
	impressionClient impression_service.ImpressionServiceClient // Client pour le service d'impression
	ad_service.UnimplementedAdServiceServer
}

// NewAdHandler crée une nouvelle instance du handler
func NewAdHandler(adService in.AdService, impressionClient impression_service.ImpressionServiceClient) *AdHandler {
	return &AdHandler{
		adService:        adService,
		impressionClient: impressionClient,
	}
}

// CreateAd implémente la création d'une publicité
func (h *AdHandler) CreateAd(ctx context.Context, req *ad_service.CreateAdRequest) (*ad_service.AdResponse, error) {
	start := time.Now()
	log.Printf("[CreateAd] start: title=%q expiresAt=%v", req.Title, req.ExpiresAt)

	// Validation des entrées
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if len(req.Title) > 100 {
		return nil, status.Error(codes.InvalidArgument, "title must be at most 100 characters")
	}
	if len(req.Description) > 500 {
		return nil, status.Error(codes.InvalidArgument, "description must be at most 500 characters")
	}

	// Transformation en objet domaine
	ad := &domain.Pub{
		Title:       req.Title,
		Description: &req.Description,
	}
	if req.ExpiresAt != nil {
		ad.ExpiresAt = req.ExpiresAt.AsTime()
	}

	// Appel au service
	createdAd, err := h.adService.CreateAd(ctx, ad)
	if err != nil {
		log.Printf("[CreateAd] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Transformation en réponse
	resp := &ad_service.AdResponse{
		Id:          createdAd.ID.String(),
		Title:       createdAd.Title,
		Description: *createdAd.Description,
		Url:         createdAd.URL,
		ExpiresAt:   timestamppb.New(createdAd.ExpiresAt),
	}
	log.Printf("[CreateAd] completed in %v id=%s", time.Since(start), createdAd.ID)
	return resp, nil
}

// GetAd implémente la récupération d'une publicité
func (h *AdHandler) GetAd(ctx context.Context, req *ad_service.GetAdRequest) (*ad_service.AdResponse, error) {
	start := time.Now()
	log.Printf("[GetAd] start: id=%q", req.Id)

	// Validation des entrées
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Appel au service
	ad, err := h.adService.GetAd(ctx, req.Id)
	if err != nil {
		log.Printf("[GetAd] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Transformation en réponse
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

	// Validation des entrées
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id format")
	}

	// Appel au service local
	url, impressions, err := h.adService.ServeAd(ctx, id)
	if err != nil {
		log.Printf("[ServeAd] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Appel au service d'impression via gRPC
	impressionID := uuid.New().String()
	trackReq := &impression_service.TrackImpressionRequest{
		AdId:         req.Id,
		ImpressionId: impressionID,
	}

	_, err = h.impressionClient.TrackImpression(ctx, trackReq)
	if err != nil {
		// On log l'erreur mais on continue pour retourner l'URL
		log.Printf("[ServeAd] impression tracking error: %v", err)
	} else {
		log.Printf("[ServeAd] impression tracked successfully: adId=%s impressionId=%s", req.Id, impressionID)
	}

	// Transformation en réponse
	resp := &ad_service.ServeAdResponse{
		Url:         url,
		Impressions: impressions,
	}
	log.Printf("[ServeAd] completed in %v id=%s impressions=%d", time.Since(start), req.Id, impressions)
	return resp, nil
}

// GetImpressionCount implémente la récupération du nombre d'impressions d'une publicité
func (h *AdHandler) GetImpressionCount(ctx context.Context, req *ad_service.GetImpressionCountRequest) (*ad_service.GetImpressionCountResponse, error) {
	start := time.Now()
	log.Printf("[GetImpressionCount] start: adId=%q", req.AdId)

	// Validation des entrées
	if req.AdId == "" {
		return nil, status.Error(codes.InvalidArgument, "ad_id is required")
	}

	id, err := uuid.Parse(req.AdId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ad_id format")
	}

	// Appel au service
	impr, err := h.adService.GetAdImpressions(ctx, id)
	if err != nil {
		log.Printf("[GetImpressionCount] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Transformation en réponse
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

// DeleteExpired implémente la suppression des annonces expirées
func (h *AdHandler) DeleteExpired(ctx context.Context, req *ad_service.DeleteExpiredRequest) (*ad_service.DeleteExpiredResponse, error) {
	start := time.Now()
	log.Printf("[DeleteExpired] start")

	// Appel au service
	count, err := h.adService.DeleteExpired(ctx)
	if err != nil {
		log.Printf("[DeleteExpired] service error: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Transformation en réponse
	resp := &ad_service.DeleteExpiredResponse{DeletedCount: count}
	log.Printf("[DeleteExpired] completed in %v deletedCount=%d", time.Since(start), count)
	return resp, nil
}
