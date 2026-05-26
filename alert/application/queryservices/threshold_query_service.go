package queryservices

import (
	"context"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

type ThresholdQueryService struct {
	repo outboundservices.AlertThresholdRepository
}

func NewThresholdQueryService(repo outboundservices.AlertThresholdRepository) *ThresholdQueryService {
	return &ThresholdQueryService{repo: repo}
}

func (s *ThresholdQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.AlertThreshold, error) {
	return s.repo.ListByUser(ctx, userID)
}
