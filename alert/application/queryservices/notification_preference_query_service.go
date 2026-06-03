// notification_preference_query_service.go — Query Service (read side, CQRS)
// for the NotificationPreference aggregate.

package queryservices

import (
	"context"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

// NotificationPreferenceQueryService answers read queries about preferences.
type NotificationPreferenceQueryService struct {
	repo outboundservices.NotificationPreferenceRepository // read-only repository use
}

// NewNotificationPreferenceQueryService wires the injected repository.
func NewNotificationPreferenceQueryService(repo outboundservices.NotificationPreferenceRepository) *NotificationPreferenceQueryService {
	return &NotificationPreferenceQueryService{repo: repo}
}

// ListByUser returns all notification preferences that belong to a given user.
func (s *NotificationPreferenceQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.NotificationPreference, error) {
	return s.repo.ListByUser(ctx, userID)
}
