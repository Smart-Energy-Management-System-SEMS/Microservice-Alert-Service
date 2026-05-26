package queryservices

import (
    "context"

    "github.com/google/uuid"

    "microservice-alert-service/alert/application/outboundservices"
    "microservice-alert-service/alert/domain/model/entities"
)

type NotificationPreferenceQueryService struct {
    repo outboundservices.NotificationPreferenceRepository
}

func NewNotificationPreferenceQueryService(repo outboundservices.NotificationPreferenceRepository) *NotificationPreferenceQueryService {
    return &NotificationPreferenceQueryService{repo: repo}
}

func (s *NotificationPreferenceQueryService) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.NotificationPreference, error) {
    return s.repo.ListByUser(ctx, userID)
}
