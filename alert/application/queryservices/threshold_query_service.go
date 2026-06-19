// threshold_query_service.go — Query Service (read side, CQRS) for the
// AlertThreshold aggregate.

package queryservices

import (
	"context"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

// ThresholdQueryService answers read queries about alert thresholds.
type ThresholdQueryService struct {
	repo outboundservices.AlertThresholdRepository // read-only repository use
}

// NewThresholdQueryService wires the injected repository.
func NewThresholdQueryService(repo outboundservices.AlertThresholdRepository) *ThresholdQueryService {
	return &ThresholdQueryService{repo: repo}
}

// ListByUser returns all thresholds that belong to a given user.
func (s *ThresholdQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.AlertThreshold, error) {
	return s.repo.ListByUser(ctx, userID)
}
