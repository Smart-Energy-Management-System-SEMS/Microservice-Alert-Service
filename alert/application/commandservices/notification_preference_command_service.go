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

type NotificationPreferenceCommandService struct {
	repo   outboundservices.NotificationPreferenceRepository
	logger *log.Logger
}

func NewNotificationPreferenceCommandService(repo outboundservices.NotificationPreferenceRepository, logger *log.Logger) *NotificationPreferenceCommandService {
	return &NotificationPreferenceCommandService{repo: repo, logger: logger}
}

func (s *NotificationPreferenceCommandService) CreatePreference(ctx context.Context, cmd commands.CreateNotificationPreferenceCommand) (*entities.NotificationPreference, error) {
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

	if err := s.repo.Create(ctx, preference); err != nil {
		return nil, err
	}

	return preference, nil
}
