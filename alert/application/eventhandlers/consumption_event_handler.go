package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/commandservices"
	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/services"
)

type IntegrationEventHandler struct {
	thresholdRepo  outboundservices.AlertThresholdRepository
	inactivityRepo outboundservices.InactivityRuleRepository
	activityRepo   outboundservices.DeviceActivityRepository
	alertService   *commandservices.AlertCommandService
	logger         *log.Logger
}

type topicAlertDefinition struct {
	AlertType string
	Title     string
	Severity  string
}

func NewIntegrationEventHandler(
	thresholdRepo outboundservices.AlertThresholdRepository,
	inactivityRepo outboundservices.InactivityRuleRepository,
	activityRepo outboundservices.DeviceActivityRepository,
	alertService *commandservices.AlertCommandService,
	logger *log.Logger,
) *IntegrationEventHandler {
	return &IntegrationEventHandler{
		thresholdRepo:  thresholdRepo,
		inactivityRepo: inactivityRepo,
		activityRepo:   activityRepo,
		alertService:   alertService,
		logger:         logger,
	}
}

func (h *IntegrationEventHandler) HandleMessage(ctx context.Context, topic string, payload []byte) error {
	switch strings.ToLower(strings.TrimSpace(topic)) {
	case "energy.consumption.recorded", "energy.reading.created":
		return h.handleEnergyEvent(ctx, topic, payload)
	default:
		return h.handleGenericEvent(ctx, topic, payload)
	}
}

func (h *IntegrationEventHandler) handleEnergyEvent(ctx context.Context, topic string, payload []byte) error {
	body, err := parsePayload(payload)
	if err != nil {
		return err
	}

	userID, ok := extractUUID(body, "user_id", "userId", "owner_id", "customer_id")
	if !ok {
		h.logger.Printf("energy event skipped topic=%s: missing user_id", topic)
		return nil
	}

	deviceID, ok := extractUUID(body, "device_id", "deviceId")
	if !ok {
		h.logger.Printf("energy event skipped topic=%s: missing device_id", topic)
		return nil
	}

	recordedAt := extractTime(body, "recorded_at", "timestamp", "created_at", "triggered_at")
	metrics := extractEnergyMetrics(body)
	if len(metrics) == 0 {
		h.logger.Printf("energy event skipped topic=%s: no supported metrics", topic)
		return nil
	}

	if err := h.evaluateThresholds(ctx, userID, deviceID, metrics, recordedAt, topic); err != nil {
		return err
	}

	if err := h.evaluateInactivity(ctx, userID, deviceID, recordedAt); err != nil {
		return err
	}

	if err := h.activityRepo.SaveLastActivity(ctx, deviceID, recordedAt); err != nil {
		h.logger.Printf("activity save error topic=%s: %v", topic, err)
	}

	return nil
}

func (h *IntegrationEventHandler) handleGenericEvent(ctx context.Context, topic string, payload []byte) error {
	definition, supported := alertDefinitions()[strings.ToLower(strings.TrimSpace(topic))]
	if !supported {
		h.logger.Printf("topic consumed without alert mapping: %s", topic)
		return nil
	}

	body, err := parsePayload(payload)
	if err != nil {
		return err
	}

	userID, ok := extractUUID(body, "user_id", "userId", "owner_id", "customer_id", "account_id")
	if !ok {
		h.logger.Printf("generic event skipped topic=%s: missing user_id", topic)
		return nil
	}

	deviceID, _ := extractUUID(body, "device_id", "deviceId")
	triggeredAt := extractTime(body, "triggered_at", "timestamp", "created_at", "processed_at", "event_time", "occurred_at")

	if definition.AlertType == "device_status" {
		definition.Severity = deriveDeviceStatusSeverity(body, definition.Severity)
	}

	if definition.AlertType == "analytics_anomaly" {
		definition.Severity = firstNonEmptyString(body, []string{"severity", "alert_severity"}, definition.Severity)
	}

	cmd := commands.CreateAlertCommand{
		UserID:      userID,
		DeviceID:    deviceID,
		AlertType:   definition.AlertType,
		Title:       firstNonEmptyString(body, []string{"title", "event_title"}, definition.Title),
		Message:     buildGenericMessage(topic, body, definition.Title),
		Severity:    normalizeSeverity(firstNonEmptyString(body, []string{"severity", "alert_severity"}, definition.Severity)),
		Status:      "open",
		TriggeredAt: triggeredAt,
	}

	if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
		h.logger.Printf("generic alert creation error topic=%s: %v", topic, err)
	}

	return nil
}

func (h *IntegrationEventHandler) evaluateThresholds(
	ctx context.Context,
	userID uuid.UUID,
	deviceID uuid.UUID,
	metrics map[string]float64,
	recordedAt time.Time,
	topic string,
) error {
	thresholds, err := h.thresholdRepo.ListActiveByUserDevice(ctx, userID, deviceID)
	if err != nil {
		return err
	}

	for _, threshold := range thresholds {
		value, found := metricLookup(metrics, threshold.Metric)
		if !found {
			continue
		}

		triggered, evalErr := services.EvaluateThreshold(threshold.Operator, value, threshold.ThresholdValue)
		if evalErr != nil {
			h.logger.Printf("threshold evaluation error topic=%s metric=%s: %v", topic, threshold.Metric, evalErr)
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
			Message:     buildThresholdMessage(threshold.Metric, value, threshold.ThresholdValue, threshold.Operator, topic),
			Severity:    "high",
			Status:      "open",
			TriggeredAt: recordedAt,
		}

		if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
			h.logger.Printf("alert creation error topic=%s: %v", topic, err)
		}
	}

	return nil
}

func (h *IntegrationEventHandler) evaluateInactivity(
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
		if !services.IsInactive(lastActive, recordedAt, rule.MaxInactiveMinutes) {
			continue
		}

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

	return nil
}

func parsePayload(payload []byte) (map[string]any, error) {
	body := map[string]any{}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, err
	}

	if data, ok := body["data"].(map[string]any); ok {
		return data, nil
	}

	return body, nil
}

func extractUUID(values map[string]any, keys ...string) (uuid.UUID, bool) {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}

		text, ok := value.(string)
		if !ok {
			continue
		}

		parsed, err := uuid.Parse(strings.TrimSpace(text))
		if err == nil {
			return parsed, true
		}
	}

	return uuid.Nil, false
}

func extractTime(values map[string]any, keys ...string) time.Time {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}

		text, ok := value.(string)
		if !ok {
			continue
		}

		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if parsed, err := time.Parse(layout, strings.TrimSpace(text)); err == nil {
				return parsed
			}
		}
	}

	return time.Now().UTC()
}

func extractEnergyMetrics(values map[string]any) map[string]float64 {
	metrics := make(map[string]float64)
	for _, key := range []string{"power_watts", "energy_kwh", "estimated_cost", "voltage", "current", "frequency", "value"} {
		if value, ok := extractFloat(values, key); ok {
			name := key
			if key == "value" {
				name = firstNonEmptyString(values, []string{"metric"}, "value")
			}
			metrics[strings.ToLower(strings.TrimSpace(name))] = value
		}
	}

	return metrics
}

func extractFloat(values map[string]any, key string) (float64, bool) {
	value, ok := values[key]
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		if err == nil {
			return parsed, true
		}
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed, true
		}
	}

	return 0, false
}

func metricLookup(metrics map[string]float64, metric string) (float64, bool) {
	value, ok := metrics[strings.ToLower(strings.TrimSpace(metric))]
	return value, ok
}

func buildThresholdMessage(metric string, value float64, thresholdValue float64, operator string, topic string) string {
	return fmt.Sprintf(
		"Metric %s value %s %s %s from topic %s",
		metric,
		formatFloat(value),
		operator,
		formatFloat(thresholdValue),
		topic,
	)
}

func buildGenericMessage(topic string, values map[string]any, fallbackTitle string) string {
	if message := firstNonEmptyString(values, []string{"message", "description", "details", "reason"}, ""); message != "" {
		return message
	}

	summaryFields := []string{
		"status",
		"result",
		"anomaly_type",
		"recommendation",
		"plan_name",
		"device_name",
		"provider",
		"currency",
	}
	summaryParts := make([]string, 0, len(summaryFields))
	for _, field := range summaryFields {
		if value := stringify(values[field]); value != "" {
			summaryParts = append(summaryParts, fmt.Sprintf("%s=%s", field, value))
		}
	}
	sort.Strings(summaryParts)
	if len(summaryParts) > 0 {
		return fmt.Sprintf("%s (%s)", fallbackTitle, strings.Join(summaryParts, ", "))
	}

	return fmt.Sprintf("%s received from topic %s", fallbackTitle, topic)
}

func stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return formatFloat(typed)
	case float32:
		return formatFloat(float64(typed))
	case int:
		return strconv.Itoa(typed)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func firstNonEmptyString(values map[string]any, keys []string, fallback string) string {
	for _, key := range keys {
		if text := stringify(values[key]); text != "" {
			return text
		}
	}

	return fallback
}

func deriveDeviceStatusSeverity(values map[string]any, fallback string) string {
	status := strings.ToLower(firstNonEmptyString(values, []string{"status", "device_status"}, ""))
	switch status {
	case "offline", "disconnected", "error", "failed":
		return "high"
	case "warning", "degraded":
		return "medium"
	case "":
		return fallback
	default:
		return "low"
	}
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "critical", "high", "medium", "low":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "low"
	}
}

func alertDefinitions() map[string]topicAlertDefinition {
	return map[string]topicAlertDefinition{
		"analytics.anomaly.detected":              {AlertType: "analytics_anomaly", Title: "Analytics anomaly detected", Severity: "critical"},
		"analytics.bill_prediction.generated":     {AlertType: "bill_prediction", Title: "Bill prediction generated", Severity: "medium"},
		"analytics.consumption_ranking.generated": {AlertType: "consumption_ranking", Title: "Consumption ranking updated", Severity: "low"},
		"analytics.device_identified":             {AlertType: "device_identified", Title: "Device identified by analytics", Severity: "low"},
		"analytics.recommendation.generated":      {AlertType: "analytics_recommendation", Title: "Analytics recommendation generated", Severity: "low"},
		"device.configuration.updated":            {AlertType: "device_configuration", Title: "Device configuration updated", Severity: "low"},
		"device.event.recorded":                   {AlertType: "device_event", Title: "Device event recorded", Severity: "low"},
		"device.linked":                           {AlertType: "device_linked", Title: "Device linked", Severity: "low"},
		"device.registered":                       {AlertType: "device_registered", Title: "Device registered", Severity: "low"},
		"device.status.updated":                   {AlertType: "device_status", Title: "Device status updated", Severity: "medium"},
		"device.unlinked":                         {AlertType: "device_unlinked", Title: "Device unlinked", Severity: "medium"},
		"iam.role-assignment.requested":           {AlertType: "role_assignment_requested", Title: "Role assignment requested", Severity: "low"},
		"iam.role.assigned":                       {AlertType: "role_assigned", Title: "Role assigned", Severity: "medium"},
		"iam.user.logged-in":                      {AlertType: "user_logged_in", Title: "User logged in", Severity: "low"},
		"iam.user.registered":                     {AlertType: "user_registered", Title: "User registered", Severity: "low"},
		"invoice.generated":                       {AlertType: "invoice_generated", Title: "Invoice generated", Severity: "medium"},
		"monitoring.alert.created":                {AlertType: "monitoring_alert", Title: "Monitoring alert created", Severity: "high"},
		"monitoring.reading.ingest":               {AlertType: "monitoring_ingest", Title: "Monitoring reading ingested", Severity: "low"},
		"monitoring.reading.processed":            {AlertType: "monitoring_processed", Title: "Monitoring reading processed", Severity: "low"},
		"payment.failed":                          {AlertType: "payment_failed", Title: "Payment failed", Severity: "critical"},
		"payment.method.added":                    {AlertType: "payment_method_added", Title: "Payment method added", Severity: "low"},
		"payment.processed":                       {AlertType: "payment_processed", Title: "Payment processed", Severity: "low"},
		"subscription.cancelled":                  {AlertType: "subscription_cancelled", Title: "Subscription cancelled", Severity: "high"},
		"subscription.created":                    {AlertType: "subscription_created", Title: "Subscription created", Severity: "low"},
		"subscription.expired":                    {AlertType: "subscription_expired", Title: "Subscription expired", Severity: "critical"},
		"subscription.plan.changed":               {AlertType: "subscription_plan_changed", Title: "Subscription plan changed", Severity: "medium"},
		"subscription.renewal.requested":          {AlertType: "subscription_renewal_requested", Title: "Subscription renewal requested", Severity: "medium"},
		"subscription.updated":                    {AlertType: "subscription_updated", Title: "Subscription updated", Severity: "low"},
	}
}

func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(value, 'f', 4, 64), "0"), ".")
}
