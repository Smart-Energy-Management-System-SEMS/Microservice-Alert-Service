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

type NotificationService struct {
	preferenceRepo outboundservices.NotificationPreferenceRepository
	logRepo        outboundservices.NotificationLogRepository
	emailSender    outboundservices.EmailSender
	smsSender      outboundservices.SmsSender
	defaultEmailTo string
	defaultSmsTo   string
	logger         *log.Logger
}

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

func (s *NotificationService) Notify(ctx context.Context, alert *entities.Alert) error {
	preferences, err := s.preferenceRepo.ListByUser(ctx, alert.UserID)
	if err != nil {
		return err
	}

	if len(preferences) == 0 {
		s.logger.Printf("no notification preferences for user %s", alert.UserID)
		return nil
	}

	var firstErr error
	now := time.Now().UTC()
	for _, preference := range preferences {
		if !preference.Enabled {
			continue
		}

		if !severityAllowed(alert.Severity, preference.MinSeverity) {
			continue
		}

		if isQuietNow(preference, now) {
			s.logNotification(ctx, alert, preference.Channel, "", "skipped", nil, strPtr("quiet hours"))
			continue
		}

		channel := strings.ToLower(preference.Channel)
		switch channel {
		case "email":
			recipient := s.defaultEmailTo
			if recipient == "" {
				s.logNotification(ctx, alert, channel, "", "skipped", nil, strPtr("missing recipient"))
				continue
			}

			if sendErr := s.emailSender.Send(ctx, recipient, alert.Title, alert.Message); sendErr != nil {
				s.logNotification(ctx, alert, channel, recipient, "failed", nil, strPtr(sendErr.Error()))
				if firstErr == nil {
					firstErr = sendErr
				}
				continue
			}

			sentAt := time.Now().UTC()
			s.logNotification(ctx, alert, channel, recipient, "sent", &sentAt, nil)
		case "sms":
			recipient := s.defaultSmsTo
			if recipient == "" {
				s.logNotification(ctx, alert, channel, "", "skipped", nil, strPtr("missing recipient"))
				continue
			}

			body := fmt.Sprintf("%s - %s", alert.Title, alert.Message)
			if sendErr := s.smsSender.Send(ctx, recipient, body); sendErr != nil {
				s.logNotification(ctx, alert, channel, recipient, "failed", nil, strPtr(sendErr.Error()))
				if firstErr == nil {
					firstErr = sendErr
				}
				continue
			}

			sentAt := time.Now().UTC()
			s.logNotification(ctx, alert, channel, recipient, "sent", &sentAt, nil)
		default:
			s.logNotification(ctx, alert, channel, "", "skipped", nil, strPtr("unsupported channel"))
		}
	}

	return firstErr
}

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

	if err := s.logRepo.Create(ctx, logEntry); err != nil {
		s.logger.Printf("notification log error: %v", err)
	}
}

func severityAllowed(alertSeverity string, minSeverity string) bool {
	severityRank := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	alertLevel := severityRank[strings.ToLower(alertSeverity)]
	minLevel := severityRank[strings.ToLower(minSeverity)]

	if alertLevel == 0 {
		alertLevel = 1
	}

	if minLevel == 0 {
		minLevel = 1
	}

	return alertLevel >= minLevel
}

func isQuietNow(preference entities.NotificationPreference, now time.Time) bool {
	if preference.QuietHoursStart == nil || preference.QuietHoursEnd == nil {
		return false
	}

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

	if end.Before(start) {
		return now.After(start) || now.Before(end)
	}

	return now.After(start) && now.Before(end)
}

func strPtr(value string) *string {
	return &value
}
