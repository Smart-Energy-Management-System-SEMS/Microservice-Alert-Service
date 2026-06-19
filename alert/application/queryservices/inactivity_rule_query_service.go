// inactivity_rule_query_service.go — Query Service (read side, CQRS) for the
// InactivityRule aggregate.

package queryservices

import (
	"context"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

// InactivityRuleQueryService answers read queries about inactivity rules.
type InactivityRuleQueryService struct {
	repo outboundservices.InactivityRuleRepository // read-only repository use
}

// NewInactivityRuleQueryService wires the injected repository.
func NewInactivityRuleQueryService(repo outboundservices.InactivityRuleRepository) *InactivityRuleQueryService {
	return &InactivityRuleQueryService{repo: repo}
}

// ListByUser returns all inactivity rules that belong to a given user.
func (s *InactivityRuleQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.InactivityRule, error) {
	return s.repo.ListByUser(ctx, userID)
}
