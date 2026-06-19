// notification_service.go — Application service that delivers an alert to the
// user across the configured channels (email/SMS), honoring their preferences
// (enabled flag, minimum severity, quiet hours) and logging every attempt.

package commandservices

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/entities"
)

// NotificationService fans out a single alert to the user's notification
// channels. It depends only on outbound abstractions (repositories and the
// email/SMS senders), keeping it independent from concrete providers.
type NotificationService struct {
	preferenceRepo outboundservices.NotificationPreferenceRepository // user channel preferences
	logRepo        outboundservices.NotificationLogRepository        // audit log of attempts
	emailSender    outboundservices.EmailSender                      // email transport port
	smsSender      outboundservices.SmsSender                        // SMS transport port
	defaultEmailTo string                                            // fallback email recipient
	defaultSmsTo   string                                            // fallback SMS recipient
	logger         *log.Logger
}

// NewNotificationService wires all injected dependencies for the service.
func NewNotificationService(
	preferenceRepo outboundservices.NotificationPreferenceRepository,
	logRepo outboundservices.NotificationLogRepository,
	emailSender outboundservices.EmailSender,
	smsSender outboundservices.SmsSender,
	defaultEmailTo string,
	defaultSmsTo string,
	logger *log.Logger,
) *NotificationService {
	return &NotificationService{
		preferenceRepo: preferenceRepo,
		logRepo:        logRepo,
		emailSender:    emailSender,
		smsSender:      smsSender,
		defaultEmailTo: defaultEmailTo,
		defaultSmsTo:   defaultSmsTo,
		logger:         logger,
	}
}

// Notify sends the alert through each eligible preference of the user.
// It returns the first delivery error encountered (if any) while still
// attempting the remaining channels.
func (s *NotificationService) Notify(ctx context.Context, alert *entities.Alert) error {
	// Load all notification preferences configured by the alert's user.
	preferences, err := s.preferenceRepo.ListByUser(ctx, alert.UserID)
	if err != nil {
		return err
	}

	// No preferences -> nothing to send (not an error).
	if len(preferences) == 0 {
		s.logger.Printf("no notification preferences for user %s", alert.UserID)
		return nil
	}

	var firstErr error
	now := time.Now().UTC()
	for _, preference := range preferences {
		// Skip channels the user disabled.
		if !preference.Enabled {
			continue
		}

		// Skip if the alert severity is below the user's minimum threshold.
		if !severityAllowed(alert.Severity, preference.MinSeverity) {
			continue
		}

		// Skip (and log) if we are currently within the user's quiet hours.
		if isQuietNow(preference, now) {
			s.logNotification(ctx, alert, preference.Channel, "", "skipped", nil, strPtr("quiet hours"))
			continue
		}

		channel := strings.ToLower(preference.Channel)
		switch channel {
		case "email":
			// Resolve the recipient; skip if none is configured.
			recipient := s.defaultEmailTo
			if recipient == "" {
				s.logNotification(ctx, alert, channel, "", "skipped", nil, strPtr("missing recipient"))
				continue
			}

			// Attempt delivery; on failure record it and remember the first error.
			if sendErr := s.emailSender.Send(ctx, recipient, alert.Title, alert.Message); sendErr != nil {
				s.logNotification(ctx, alert, channel, recipient, "failed", nil, strPtr(sendErr.Error()))
				if firstErr == nil {
					firstErr = sendErr
				}
				continue
			}

			// Success: log the sent timestamp.
			sentAt := time.Now().UTC()
			s.logNotification(ctx, alert, channel, recipient, "sent", &sentAt, nil)
		case "sms":
			// Resolve the recipient; skip if none is configured.
			recipient := s.defaultSmsTo
			if recipient == "" {
				s.logNotification(ctx, alert, channel, "", "skipped", nil, strPtr("missing recipient"))
				continue
			}

			// Compose the SMS body and attempt delivery.
			body := fmt.Sprintf("%s - %s", alert.Title, alert.Message)
			if sendErr := s.smsSender.Send(ctx, recipient, body); sendErr != nil {
				s.logNotification(ctx, alert, channel, recipient, "failed", nil, strPtr(sendErr.Error()))
				if firstErr == nil {
					firstErr = sendErr
				}
				continue
			}

			// Success: log the sent timestamp.
			sentAt := time.Now().UTC()
			s.logNotification(ctx, alert, channel, recipient, "sent", &sentAt, nil)
		default:
			// Unknown channel: record it as skipped.
			s.logNotification(ctx, alert, channel, "", "skipped", nil, strPtr("unsupported channel"))
		}
	}

	return firstErr
}

// logNotification persists one entry in the notification audit log,
// capturing the outcome (sent/failed/skipped) of a delivery attempt.
func (s *NotificationService) logNotification(
	ctx context.Context,
	alert *entities.Alert,
	channel string,
	recipient string,
	status string,
	sentAt *time.Time,
	errMessage *string,
) {
	logEntry := &entities.NotificationLog{
		NotificationID: uuid.New(),
		AlertID:        alert.AlertID,
		Channel:        channel,
		Recipient:      recipient,
		Status:         status,
		SentAt:         sentAt,
		ErrorMessage:   errMessage,
		CreatedAt:      time.Now().UTC(),
	}

	// A logging failure must not break the notification flow; just report it.
	if err := s.logRepo.Create(ctx, logEntry); err != nil {
		s.logger.Printf("notification log error: %v", err)
	}
}

// severityAllowed reports whether an alert's severity reaches the user's
// minimum severity, using a numeric ranking. Unknown values default to "low".
func severityAllowed(alertSeverity string, minSeverity string) bool {
	severityRank := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	alertLevel := severityRank[strings.ToLower(alertSeverity)]
	minLevel := severityRank[strings.ToLower(minSeverity)]

	// Default unknown severities to the lowest rank.
	if alertLevel == 0 {
		alertLevel = 1
	}

	if minLevel == 0 {
		minLevel = 1
	}

	return alertLevel >= minLevel
}

// isQuietNow reports whether "now" falls inside the user's quiet-hours window.
// It correctly handles windows that wrap past midnight (end before start).
func isQuietNow(preference entities.NotificationPreference, now time.Time) bool {
	// No quiet hours configured -> never quiet.
	if preference.QuietHoursStart == nil || preference.QuietHoursEnd == nil {
		return false
	}

	// Anchor the configured start/end times to today's date.
	start := time.Date(
		now.Year(), now.Month(), now.Day(),
		preference.QuietHoursStart.Hour(), preference.QuietHoursStart.Minute(), 0, 0,
		now.Location(),
	)

	end := time.Date(
		now.Year(), now.Month(), now.Day(),
		preference.QuietHoursEnd.Hour(), preference.QuietHoursEnd.Minute(), 0, 0,
		now.Location(),
	)

	// Window crosses midnight (e.g. 22:00–06:00): quiet if before end or after start.
	if end.Before(start) {
		return now.After(start) || now.Before(end)
	}

	// Normal same-day window: quiet if strictly between start and end.
	return now.After(start) && now.Before(end)
}

// strPtr is a small helper that returns a pointer to the given string,
// used to pass optional string fields by reference.
func strPtr(value string) *string {
	return &value
}
