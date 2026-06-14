// Package commandservices holds the Application-layer Command Services
// (the "write" side of the CQRS pattern). These services orchestrate the
// domain, persist changes through repository abstractions, and trigger
// side effects such as notifications.
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

// AlertCommandService coordinates the creation of alerts and the update of
// their status. It depends on abstractions (repository and notifier) so it
// stays decoupled from concrete infrastructure (Dependency Inversion).
type AlertCommandService struct {
	repo      outboundservices.AlertRepository
	publisher outboundservices.AlertEventPublisher
	notifier  *NotificationService
	logger    *log.Logger
}

// NewAlertCommandService is the constructor; it wires the injected dependencies.
func NewAlertCommandService(
	repo outboundservices.AlertRepository,
	publisher outboundservices.AlertEventPublisher,
	notifier *NotificationService,
	logger *log.Logger,
) *AlertCommandService {
	return &AlertCommandService{repo: repo, publisher: publisher, notifier: notifier, logger: logger}
}

// CreateAlert builds an Alert entity from the command and persists it.
func (s *AlertCommandService) CreateAlert(ctx context.Context, cmd commands.CreateAlertCommand) (*entities.Alert, error) {
	// Map the input command into a domain entity, assigning a fresh UUID.
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

	// Default the trigger time to "now" when the caller did not provide one.
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now().UTC()
	}

	// Persist; on failure propagate the error to the caller.
	if err := s.repo.Create(ctx, alert); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishJSON(ctx, alert.AlertID.String(), buildAlertCreatedEvent(alert)); err != nil {
			s.logger.Printf("alert event publish error: %v", err)
			return alert, err
		}
	}

	return alert, nil
}

// CreateAlertAndNotify persists the alert and then dispatches notifications.
// The alert is returned even if notification fails, alongside the error.
func (s *AlertCommandService) CreateAlertAndNotify(ctx context.Context, cmd commands.CreateAlertCommand) (*entities.Alert, error) {
	alert, err := s.CreateAlert(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Notify only if a notifier is configured (it is optional).
	if s.notifier != nil {
		if notifyErr := s.notifier.Notify(ctx, alert); notifyErr != nil {
			s.logger.Printf("notification error: %v", notifyErr)
			return alert, notifyErr
		}
	}

	return alert, nil
}

// UpdateAlertStatus changes the status of an existing alert (e.g. resolved).
func (s *AlertCommandService) UpdateAlertStatus(ctx context.Context, cmd commands.UpdateAlertStatusCommand) error {
	return s.repo.UpdateStatus(ctx, cmd.AlertID, cmd.Status, cmd.ResolvedAt)
}

func buildAlertCreatedEvent(alert *entities.Alert) map[string]any {
	data := map[string]any{
		"alert_id":     alert.AlertID.String(),
		"user_id":      alert.UserID.String(),
		"device_id":    alert.DeviceID.String(),
		"alert_type":   alert.AlertType,
		"title":        alert.Title,
		"message":      alert.Message,
		"severity":     alert.Severity,
		"status":       alert.Status,
		"triggered_at": alert.TriggeredAt,
	}

	if alert.ThresholdID != nil {
		data["threshold_id"] = alert.ThresholdID.String()
	}

	if alert.InactivityRuleID != nil {
		data["inactivity_rule_id"] = alert.InactivityRuleID.String()
	}

	if alert.ResolvedAt != nil {
		data["resolved_at"] = alert.ResolvedAt
	}

	event := map[string]any{
		"eventType":  "alert.created",
		"event":      "alert.created",
		"source":     "alert-service",
		"occurredAt": alert.TriggeredAt.UTC().Format(time.RFC3339),
		"data":       data,
	}

	for key, value := range data {
		event[key] = value
	}

	return event
}
