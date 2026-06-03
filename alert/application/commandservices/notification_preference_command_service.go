// notification_preference_command_service.go — Command Service (write side,
// CQRS) for the NotificationPreference aggregate: stores how and when a user
// wants to be notified (channel, minimum severity, quiet hours).

package commandservices

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/model/entities"
)

// NotificationPreferenceCommandService orchestrates the creation of preferences.
type NotificationPreferenceCommandService struct {
	repo   outboundservices.NotificationPreferenceRepository // persistence contract
	logger *log.Logger
}

// NewNotificationPreferenceCommandService wires the injected dependencies.
func NewNotificationPreferenceCommandService(repo outboundservices.NotificationPreferenceRepository, logger *log.Logger) *NotificationPreferenceCommandService {
	return &NotificationPreferenceCommandService{repo: repo, logger: logger}
}

// CreatePreference builds a NotificationPreference entity and persists it.
func (s *NotificationPreferenceCommandService) CreatePreference(ctx context.Context, cmd commands.CreateNotificationPreferenceCommand) (*entities.NotificationPreference, error) {
	// Map the command to a domain entity with a fresh id and UTC timestamps.
	preference := &entities.NotificationPreference{
		PreferenceID:    uuid.New(),
		UserID:          cmd.UserID,
		Channel:         cmd.Channel,
		Enabled:         cmd.Enabled,
		MinSeverity:     cmd.MinSeverity,
		QuietHoursStart: cmd.QuietHoursStart,
		QuietHoursEnd:   cmd.QuietHoursEnd,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	// Persist; propagate any error to the caller.
	if err := s.repo.Create(ctx, preference); err != nil {
		return nil, err
	}

	return preference, nil
}
