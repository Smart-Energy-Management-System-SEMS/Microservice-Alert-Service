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

type AlertCommandService struct {
	repo     outboundservices.AlertRepository
	notifier *NotificationService
	logger   *log.Logger
}

func NewAlertCommandService(repo outboundservices.AlertRepository, notifier *NotificationService, logger *log.Logger) *AlertCommandService {
	return &AlertCommandService{repo: repo, notifier: notifier, logger: logger}
}

func (s *AlertCommandService) CreateAlert(ctx context.Context, cmd commands.CreateAlertCommand) (*entities.Alert, error) {
	alert := &entities.Alert{
		AlertID:          uuid.New(),
		UserID:           cmd.UserID,
		DeviceID:         cmd.DeviceID,
		ThresholdID:      cmd.ThresholdID,
		InactivityRuleID: cmd.InactivityRuleID,
		AlertType:        cmd.AlertType,
		Title:            cmd.Title,
		Message:          cmd.Message,
		Severity:         cmd.Severity,
		Status:           cmd.Status,
		TriggeredAt:      cmd.TriggeredAt,
	}

	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now().UTC()
	}

	if err := s.repo.Create(ctx, alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *AlertCommandService) CreateAlertAndNotify(ctx context.Context, cmd commands.CreateAlertCommand) (*entities.Alert, error) {
	alert, err := s.CreateAlert(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.notifier != nil {
		if notifyErr := s.notifier.Notify(ctx, alert); notifyErr != nil {
			s.logger.Printf("notification error: %v", notifyErr)
			return alert, notifyErr
		}
	}

	return alert, nil
}

func (s *AlertCommandService) UpdateAlertStatus(ctx context.Context, cmd commands.UpdateAlertStatusCommand) error {
	return s.repo.UpdateStatus(ctx, cmd.AlertID, cmd.Status, cmd.ResolvedAt)
}
