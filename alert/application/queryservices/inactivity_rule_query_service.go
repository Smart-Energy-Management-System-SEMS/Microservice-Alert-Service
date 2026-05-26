package queryservices

import (
	"context"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

type InactivityRuleQueryService struct {
	repo outboundservices.InactivityRuleRepository
}

func NewInactivityRuleQueryService(repo outboundservices.InactivityRuleRepository) *InactivityRuleQueryService {
	return &InactivityRuleQueryService{repo: repo}
}

func (s *InactivityRuleQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.InactivityRule, error) {
	return s.repo.ListByUser(ctx, userID)
}
