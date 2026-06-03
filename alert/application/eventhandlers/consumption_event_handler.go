// Package eventhandlers contains the Application-layer handlers that react to
// integration events received from other microservices (via Kafka). A handler
// translates an external event into domain operations (evaluating rules and
// creating alerts), without holding business logic of its own.
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

// ConsumptionRecordedEvent is the external payload published when a device
// records a consumption metric. The json tags map the broker message fields.
type ConsumptionRecordedEvent struct {
	UserID     string  `json:"user_id"`
	DeviceID   string  `json:"device_id"`
	Metric     string  `json:"metric"`
	Value      float64 `json:"value"`
	RecordedAt string  `json:"recorded_at"`
}

// ConsumptionEventHandler reacts to consumption events by evaluating the
// user's thresholds and inactivity rules, and raising alerts when triggered.
type ConsumptionEventHandler struct {
	thresholdRepo  outboundservices.AlertThresholdRepository  // active thresholds lookup
	inactivityRepo outboundservices.InactivityRuleRepository  // active inactivity rules lookup
	activityRepo   outboundservices.DeviceActivityRepository  // last-activity tracking
	alertService   *commandservices.AlertCommandService       // creates + notifies alerts
	logger         *log.Logger
}

// NewConsumptionEventHandler wires all injected dependencies.
func NewConsumptionEventHandler(
	thresholdRepo outboundservices.AlertThresholdRepository,
	inactivityRepo outboundservices.InactivityRuleRepository,
	activityRepo outboundservices.DeviceActivityRepository,
	alertService *commandservices.AlertCommandService,
	logger *log.Logger,
) *ConsumptionEventHandler {
	return &ConsumptionEventHandler{
		thresholdRepo:  thresholdRepo,
		inactivityRepo: inactivityRepo,
		activityRepo:   activityRepo,
		alertService:   alertService,
		logger:         logger,
	}
}

// HandleMessage is the entry point for a raw Kafka message. It decodes and
// validates the payload, then runs threshold and inactivity evaluation and
// records the device's latest activity timestamp.
func (h *ConsumptionEventHandler) HandleMessage(ctx context.Context, payload []byte) error {
	// Decode the JSON payload into the event struct.
	var event ConsumptionRecordedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	// Parse and validate the incoming string fields up front.
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

	// Run the two independent rule evaluations.
	if err := h.evaluateThresholds(ctx, userID, deviceID, event, recordedAt); err != nil {
		return err
	}

	if err := h.evaluateInactivity(ctx, userID, deviceID, recordedAt); err != nil {
		return err
	}

	// Record this event as the device's most recent activity. A failure here
	// is logged but does not fail the whole message handling.
	if err := h.activityRepo.SaveLastActivity(ctx, deviceID, recordedAt); err != nil {
		h.logger.Printf("activity save error: %v", err)
	}

	return nil
}

// evaluateThresholds checks every active threshold for the device and raises
// an alert for each one that the recorded value triggers.
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
		// Only thresholds defined for this event's metric are relevant.
		if !strings.EqualFold(threshold.Metric, event.Metric) {
			continue
		}

		// Ask the domain service whether the value breaches the threshold.
		triggered, evalErr := services.EvaluateThreshold(threshold.Operator, event.Value, threshold.ThresholdValue)
		if evalErr != nil {
			h.logger.Printf("threshold evaluation error: %v", evalErr)
			continue
		}

		if !triggered {
			continue
		}

		// Build the alert command describing the breached threshold.
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

		// Create + notify; log but do not abort on individual alert failures.
		if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
			h.logger.Printf("alert creation error: %v", err)
		}
	}

	return nil
}

// evaluateInactivity raises an alert for each active inactivity rule whose
// configured limit has been exceeded since the device's last activity.
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

	// Without a known last-activity timestamp there is nothing to compare.
	lastActive, found, err := h.activityRepo.GetLastActivity(ctx, deviceID)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	for _, rule := range rules {
		// Delegate the inactivity decision to the domain service.
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

// buildThresholdMessage composes a human-readable description of the breach.
func buildThresholdMessage(event ConsumptionRecordedEvent, thresholdValue float64, operator string) string {
	return "Metric " + event.Metric + " value " + formatFloat(event.Value) + " " + operator + " " + formatFloat(thresholdValue)
}

// formatFloat renders a float without trailing zeros (e.g. 12.50 -> "12.5").
func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmtFloat(value), "0"), ".")
}

// fmtFloat formats a float with 4 decimal places before trimming.
func fmtFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64)
}
