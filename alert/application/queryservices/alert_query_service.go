// Package queryservices holds the Application-layer Query Services (the
// "read" side of the CQRS pattern). They only retrieve data through the
// repository ports; they never modify state nor emit side effects.
package queryservices

import (
	"context"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

// AlertQueryService answers read queries about alerts.
type AlertQueryService struct {
	repo outboundservices.AlertRepository // read-only use of the alert repository
}

// NewAlertQueryService wires the injected repository.
func NewAlertQueryService(repo outboundservices.AlertRepository) *AlertQueryService {
	return &AlertQueryService{repo: repo}
}

// ListAll returns every alert in the system.
func (s *AlertQueryService) ListAll(ctx context.Context) ([]entities.Alert, error) {
	return s.repo.ListAll(ctx)
}

// GetByID returns a single alert by its identifier.
func (s *AlertQueryService) GetByID(ctx context.Context, alertID uuid.UUID) (*entities.Alert, error) {
	return s.repo.GetByID(ctx, alertID)
}

// ListByUser returns all alerts that belong to a given user.
func (s *AlertQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Alert, error) {
	return s.repo.ListByUser(ctx, userID)
}
