// Package commandservices holds the Application-layer Command Services
// (the "write" side of the CQRS pattern). These services orchestrate the
// domain, persist changes through repository abstractions, and trigger
// side effects such as notifications.
package commandservices

import (
	"context"
	"log"
	"strings"
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
	repo          outboundservices.AlertRepository
	publisher     outboundservices.AlertEventPublisher
	notifier      *NotificationService
	defaultStatus string
	logger        *log.Logger
}

// NewAlertCommandService is the constructor; it wires the injected dependencies.
func NewAlertCommandService(
	repo outboundservices.AlertRepository,
	publisher outboundservices.AlertEventPublisher,
	notifier *NotificationService,
	defaultStatus string,
	logger *log.Logger,
) *AlertCommandService {
	return &AlertCommandService{
		repo:          repo,
		publisher:     publisher,
		notifier:      notifier,
		defaultStatus: normalizeAlertStatus(defaultStatus, "open"),
		logger:        logger,
	}
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
		Status:           normalizeAlertStatus(cmd.Status, s.defaultStatus),
		TriggeredAt:      cmd.TriggeredAt,
	}

	// Default the trigger time to "now" when the caller did not provide one.
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now().UTC()
	}

	s.logger.Printf(
		"create alert start alert_id=%s user_id=%s device_id=%s alert_type=%s status=%s",
		alert.AlertID,
		alert.UserID,
		alert.DeviceID,
		alert.AlertType,
		alert.Status,
	)

	// Persist; on failure propagate the error to the caller.
	if err := s.repo.Create(ctx, alert); err != nil {
		s.logger.Printf(
			"create alert persistence error alert_id=%s user_id=%s device_id=%s alert_type=%s: %v",
			alert.AlertID,
			alert.UserID,
			alert.DeviceID,
			alert.AlertType,
			err,
		)
		return nil, err
	}

	s.logger.Printf(
		"create alert persisted alert_id=%s user_id=%s device_id=%s alert_type=%s",
		alert.AlertID,
		alert.UserID,
		alert.DeviceID,
		alert.AlertType,
	)

	if s.publisher != nil {
		s.logger.Printf(
			"create alert publish attempt alert_id=%s topic=%s",
			alert.AlertID,
			s.publisher.Topic(),
		)
		if err := s.publisher.PublishJSON(ctx, alert.AlertID.String(), buildAlertCreatedEvent(alert)); err != nil {
			s.logger.Printf("alert event publish error: %v", err)
			return alert, err
		}
		s.logger.Printf(
			"create alert published alert_id=%s topic=%s",
			alert.AlertID,
			s.publisher.Topic(),
		)
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
		s.logger.Printf(
			"create alert notify attempt alert_id=%s user_id=%s device_id=%s alert_type=%s",
			alert.AlertID,
			alert.UserID,
			alert.DeviceID,
			alert.AlertType,
		)
		if notifyErr := s.notifier.Notify(ctx, alert); notifyErr != nil {
			s.logger.Printf("notification error: %v", notifyErr)
			return alert, notifyErr
		}
		s.logger.Printf(
			"create alert notify success alert_id=%s user_id=%s device_id=%s alert_type=%s",
			alert.AlertID,
			alert.UserID,
			alert.DeviceID,
			alert.AlertType,
		)
	}

	return alert, nil
}

// UpdateAlertStatus changes the status of an existing alert (e.g. resolved).
func (s *AlertCommandService) UpdateAlertStatus(ctx context.Context, cmd commands.UpdateAlertStatusCommand) error {
	return s.repo.UpdateStatus(ctx, cmd.AlertID, normalizeAlertStatus(cmd.Status, s.defaultStatus), cmd.ResolvedAt)
}

func normalizeAlertStatus(value string, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "open", "pending", "active":
		return "open"
	case "closed":
		return "resolved"
	case "resolved", "dismissed", "acknowledged":
		return strings.ToLower(strings.TrimSpace(value))
	case "":
		return strings.ToLower(strings.TrimSpace(fallback))
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
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
		"eventId":    alert.AlertID.String(),
		"occurredAt": alert.TriggeredAt.UTC().Format(time.RFC3339),
		"data":       data,
	}

	return event
}
