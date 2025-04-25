package handler

import (
	"context"
	"log"

	"impression-tracker/generated/impression_service"
	"impression-tracker/internal/ports/in"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server gère les requêtes gRPC pour les impressions
type Server struct {
	impression_service.UnimplementedImpressionServiceServer
	service in.ImpressionService
}

// NewServer crée un nouveau serveur gRPC
func NewServer(service in.ImpressionService) *Server {
	return &Server{service: service}
}

// TrackImpression enregistre une nouvelle impression pour une publicité
func (s *Server) TrackImpression(ctx context.Context, req *impression_service.TrackImpressionRequest) (*impression_service.TrackImpressionResponse, error) {
	adID := req.GetAdId()
	if adID == "" {
		return nil, status.Error(codes.InvalidArgument, "ad_id is required")
	}

	if err := s.service.Track(ctx, adID); err != nil {
		log.Printf("[TrackImpression] service error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to track impression: %v", err)
	}

	log.Printf("[TrackImpression] adID=%s incremented", adID)
	return &impression_service.TrackImpressionResponse{Success: true}, nil
}

// GetCount récupère le nombre d'impressions pour une publicité
func (s *Server) GetCount(ctx context.Context, req *impression_service.GetImpressionCountRequest) (*impression_service.GetImpressionCountResponse, error) {
	adID := req.GetAdId()
	if adID == "" {
		return nil, status.Error(codes.InvalidArgument, "ad_id is required")
	}

	count, err := s.service.GetCount(ctx, adID)
	if err != nil {
		log.Printf("[GetCount] service error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get count: %v", err)
	}

	log.Printf("[GetCount] adID=%s count=%d", adID, count)
	return &impression_service.GetImpressionCountResponse{Count: count}, nil
}
