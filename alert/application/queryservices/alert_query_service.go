package queryservices

import (
	"context"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

type AlertQueryService struct {
	repo outboundservices.AlertRepository
}

func NewAlertQueryService(repo outboundservices.AlertRepository) *AlertQueryService {
	return &AlertQueryService{repo: repo}
}

func (s *AlertQueryService) ListAll(ctx context.Context) ([]entities.Alert, error) {
	return s.repo.ListAll(ctx)
}

func (s *AlertQueryService) GetByID(ctx context.Context, alertID uuid.UUID) (*entities.Alert, error) {
	return s.repo.GetByID(ctx, alertID)
}

func (s *AlertQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Alert, error) {
	return s.repo.ListByUser(ctx, userID)
}
