package eventhandlers

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/commandservices"
	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/services"
)

type ConsumptionRecordedEvent struct {
	UserID     string  `json:"user_id"`
	DeviceID   string  `json:"device_id"`
	Metric     string  `json:"metric"`
	Value      float64 `json:"value"`
	RecordedAt string  `json:"recorded_at"`
}

type ConsumptionEventHandler struct {
	thresholdRepo   outboundservices.AlertThresholdRepository
	inactivityRepo  outboundservices.InactivityRuleRepository
	activityRepo    outboundservices.DeviceActivityRepository
	alertService    *commandservices.AlertCommandService
	logger          *log.Logger
}

func NewConsumptionEventHandler(
	thresholdRepo outboundservices.AlertThresholdRepository,
	inactivityRepo outboundservices.InactivityRuleRepository,
	activityRepo outboundservices.DeviceActivityRepository,
	alertService *commandservices.AlertCommandService,
	logger *log.Logger,
) *ConsumptionEventHandler {
	return &ConsumptionEventHandler{
		thresholdRepo:   thresholdRepo,
		inactivityRepo:  inactivityRepo,
		activityRepo:    activityRepo,
		alertService:    alertService,
		logger:          logger,
	}
}

func (h *ConsumptionEventHandler) HandleMessage(ctx context.Context, payload []byte) error {
	var event ConsumptionRecordedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	recordedAt, err := time.Parse(time.RFC3339, event.RecordedAt)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return err
	}

	deviceID, err := uuid.Parse(event.DeviceID)
	if err != nil {
		return err
	}

	if err := h.evaluateThresholds(ctx, userID, deviceID, event, recordedAt); err != nil {
		return err
	}

	if err := h.evaluateInactivity(ctx, userID, deviceID, recordedAt); err != nil {
		return err
	}

	if err := h.activityRepo.SaveLastActivity(ctx, deviceID, recordedAt); err != nil {
		h.logger.Printf("activity save error: %v", err)
	}

	return nil
}

func (h *ConsumptionEventHandler) evaluateThresholds(
	ctx context.Context,
	userID uuid.UUID,
	deviceID uuid.UUID,
	event ConsumptionRecordedEvent,
	recordedAt time.Time,
) error {
	thresholds, err := h.thresholdRepo.ListActiveByUserDevice(ctx, userID, deviceID)
	if err != nil {
		return err
	}

	for _, threshold := range thresholds {
		if !strings.EqualFold(threshold.Metric, event.Metric) {
			continue
		}

		triggered, evalErr := services.EvaluateThreshold(threshold.Operator, event.Value, threshold.ThresholdValue)
		if evalErr != nil {
			h.logger.Printf("threshold evaluation error: %v", evalErr)
			continue
		}

		if !triggered {
			continue
		}

		cmd := commands.CreateAlertCommand{
			UserID:      userID,
			DeviceID:    deviceID,
			ThresholdID: &threshold.ThresholdID,
			AlertType:   "threshold",
			Title:       threshold.ThresholdName,
			Message:     buildThresholdMessage(event, threshold.ThresholdValue, threshold.Operator),
			Severity:    "high",
			Status:      "open",
			TriggeredAt: recordedAt,
		}

		if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
			h.logger.Printf("alert creation error: %v", err)
		}
	}

	return nil
}

func (h *ConsumptionEventHandler) evaluateInactivity(
	ctx context.Context,
	userID uuid.UUID,
	deviceID uuid.UUID,
	recordedAt time.Time,
) error {
	rules, err := h.inactivityRepo.ListActiveByUserDevice(ctx, userID, deviceID)
	if err != nil {
		return err
	}

	lastActive, found, err := h.activityRepo.GetLastActivity(ctx, deviceID)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	for _, rule := range rules {
		if services.IsInactive(lastActive, recordedAt, rule.MaxInactiveMinutes) {
			cmd := commands.CreateAlertCommand{
				UserID:           userID,
				DeviceID:         deviceID,
				InactivityRuleID: &rule.InactivityRuleID,
				AlertType:        "inactivity",
				Title:            rule.RuleName,
				Message:          "Device inactive beyond configured threshold",
				Severity:         "medium",
				Status:           "open",
				TriggeredAt:      recordedAt,
			}

			if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
				h.logger.Printf("inactivity alert error: %v", err)
			}
		}
	}

	return nil
}

func buildThresholdMessage(event ConsumptionRecordedEvent, thresholdValue float64, operator string) string {
	return "Metric " + event.Metric + " value " + formatFloat(event.Value) + " " + operator + " " + formatFloat(thresholdValue)
}

func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmtFloat(value), "0"), ".")
}

func fmtFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64)
}
