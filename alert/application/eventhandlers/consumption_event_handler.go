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

type integrationEvent struct {
	Topic     string
	EventType string
	Values    map[string]any
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
	event, err := parseIntegrationEvent(topic, payload)
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(event.Topic)) {
	case "energy.events", "energy.consumption.recorded", "energy.reading.created":
		return h.handleEnergyEvent(ctx, event)
	case "analytics.events":
		return h.handleGenericEvent(ctx, event)
	default:
		return h.handleGenericEvent(ctx, event)
	}
}

func (h *IntegrationEventHandler) handleEnergyEvent(ctx context.Context, event integrationEvent) error {
	eventType := effectiveEventType(event)
	if !matchesAnyEventType(eventType, "energy.consumption.recorded", "energy.reading.created") {
		h.logDiscardedEvent(event, "unsupported energy event type")
		return nil
	}

	body := event.Values

	userID, ok := extractUUID(body, "user_id", "userId", "owner_id", "customer_id")
	if !ok {
		h.logDiscardedEvent(event, "missing user_id")
		return nil
	}

	deviceID, ok := extractUUID(body, "device_id", "deviceId")
	if !ok {
		h.logDiscardedEvent(event, "missing device_id")
		return nil
	}

	h.logger.Printf(
		"energy event accepted topic=%s eventType=%s user_id=%s device_id=%s",
		event.Topic,
		eventType,
		userID,
		deviceID,
	)

	recordedAt := extractTime(body, "recorded_at", "timestamp", "created_at", "triggered_at", "occurredAt", "occurred_at")
	metrics := extractEnergyMetrics(body)
	if len(metrics) == 0 {
		h.logDiscardedEvent(event, "no supported metrics")
		return nil
	}

	if err := h.evaluateThresholds(ctx, event, userID, deviceID, metrics, recordedAt); err != nil {
		return err
	}

	if err := h.evaluateInactivity(ctx, userID, deviceID, recordedAt); err != nil {
		return err
	}

	if err := h.activityRepo.SaveLastActivity(ctx, deviceID, recordedAt); err != nil {
		h.logger.Printf("activity save error topic=%s eventType=%s: %v", event.Topic, event.EventType, err)
	}

	return nil
}

func (h *IntegrationEventHandler) handleGenericEvent(ctx context.Context, event integrationEvent) error {
	eventType := effectiveEventType(event)
	if strings.EqualFold(strings.TrimSpace(event.Topic), "analytics.events") &&
		!matchesAnyEventType(eventType, "analytics.anomaly.detected", "analytics.recommendation.generated") {
		h.logDiscardedEvent(event, "unsupported analytics event type")
		return nil
	}

	definition, supported := alertDefinitions()[eventType]
	if !supported {
		h.logDiscardedEvent(event, "topic consumed without alert mapping")
		return nil
	}

	body := event.Values

	userID, ok := extractUUID(body, "user_id", "userId", "owner_id", "customer_id", "account_id")
	if !ok {
		h.logDiscardedEvent(event, "missing user_id")
		return nil
	}

	deviceID, _ := extractUUID(body, "device_id", "deviceId")
	triggeredAt := extractTime(body, "triggered_at", "timestamp", "created_at", "processed_at", "event_time", "occurred_at", "occurredAt")

	if definition.AlertType == "device_status" {
		definition.Severity = deriveDeviceStatusSeverity(body, definition.Severity)
	}

	if definition.AlertType == "analytics_anomaly" {
		definition.Severity = firstNonEmptyString(body, []string{"severity", "alert_severity"}, definition.Severity)
	}

	h.logger.Printf(
		"generic alert event accepted topic=%s eventType=%s user_id=%s device_id=%s alert_type=%s",
		event.Topic,
		eventType,
		userID,
		deviceID,
		definition.AlertType,
	)

	cmd := commands.CreateAlertCommand{
		UserID:      userID,
		DeviceID:    deviceID,
		AlertType:   definition.AlertType,
		Title:       firstNonEmptyString(body, []string{"title", "event_title"}, definition.Title),
		Message:     buildGenericMessage(event.Topic, eventType, body, definition.Title),
		Severity:    normalizeSeverity(firstNonEmptyString(body, []string{"severity", "alert_severity"}, definition.Severity)),
		Status:      "open",
		TriggeredAt: triggeredAt,
	}

	h.logger.Printf(
		"generic alert creation attempt topic=%s eventType=%s user_id=%s device_id=%s alert_type=%s",
		event.Topic,
		eventType,
		userID,
		deviceID,
		cmd.AlertType,
	)

	if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
		h.logger.Printf("generic alert creation error topic=%s eventType=%s: %v", event.Topic, event.EventType, err)
		return nil
	}

	h.logger.Printf(
		"generic alert persisted topic=%s eventType=%s user_id=%s device_id=%s alert_type=%s",
		event.Topic,
		eventType,
		userID,
		deviceID,
		cmd.AlertType,
	)

	return nil
}

func (h *IntegrationEventHandler) evaluateThresholds(
	ctx context.Context,
	event integrationEvent,
	userID uuid.UUID,
	deviceID uuid.UUID,
	metrics map[string]float64,
	recordedAt time.Time,
) error {
	thresholds, err := h.thresholdRepo.ListActiveByUserDevice(ctx, userID, deviceID)
	if err != nil {
		return err
	}
	if len(thresholds) == 0 {
		h.logger.Printf(
			"energy event skipped topic=%s eventType=%s user_id=%s device_id=%s reason=no active thresholds matched by user/device",
			event.Topic,
			effectiveEventType(event),
			userID,
			deviceID,
		)
		return nil
	}

	matchedMetric := false
	for _, threshold := range thresholds {
		value, found := metricLookup(metrics, threshold.Metric)
		if !found {
			continue
		}
		matchedMetric = true

		triggered, evalErr := services.EvaluateThreshold(threshold.Operator, value, threshold.ThresholdValue)
		if evalErr != nil {
			h.logger.Printf("threshold evaluation error topic=%s eventType=%s user_id=%s device_id=%s metric=%s threshold_value=%s: %v", event.Topic, effectiveEventType(event), userID, deviceID, threshold.Metric, formatFloat(threshold.ThresholdValue), evalErr)
			continue
		}
		if !triggered {
			h.logger.Printf("energy event skipped topic=%s eventType=%s user_id=%s device_id=%s reason=threshold not triggered metric=%s observed=%s threshold_value=%s", event.Topic, effectiveEventType(event), userID, deviceID, threshold.Metric, formatFloat(value), formatFloat(threshold.ThresholdValue))
			continue
		}

		cmd := commands.CreateAlertCommand{
			UserID:      userID,
			DeviceID:    deviceID,
			ThresholdID: &threshold.ThresholdID,
			AlertType:   "threshold",
			Title:       threshold.ThresholdName,
			Message:     buildThresholdMessage(threshold.Metric, value, threshold.ThresholdValue, threshold.Operator, event.Topic, effectiveEventType(event)),
			Severity:    "high",
			Status:      "open",
			TriggeredAt: recordedAt,
		}

		h.logger.Printf(
			"threshold alert creation attempt topic=%s eventType=%s user_id=%s device_id=%s metric=%s observed=%s threshold_value=%s",
			event.Topic,
			effectiveEventType(event),
			userID,
			deviceID,
			threshold.Metric,
			formatFloat(value),
			formatFloat(threshold.ThresholdValue),
		)

		if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
			h.logger.Printf("alert creation error topic=%s eventType=%s user_id=%s device_id=%s metric=%s threshold_value=%s: %v", event.Topic, effectiveEventType(event), userID, deviceID, threshold.Metric, formatFloat(threshold.ThresholdValue), err)
			continue
		}

		h.logger.Printf("alert persisted topic=%s eventType=%s user_id=%s device_id=%s metric=%s threshold_value=%s", event.Topic, effectiveEventType(event), userID, deviceID, threshold.Metric, formatFloat(threshold.ThresholdValue))
	}

	if !matchedMetric {
		h.logger.Printf("energy event skipped topic=%s eventType=%s user_id=%s device_id=%s reason=no threshold metric matched event metrics=%s", event.Topic, effectiveEventType(event), userID, deviceID, strings.Join(sortedMetricNames(metrics), ","))
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

		h.logger.Printf(
			"inactivity alert creation attempt user_id=%s device_id=%s rule_id=%s max_inactive_minutes=%d",
			userID,
			deviceID,
			rule.InactivityRuleID,
			rule.MaxInactiveMinutes,
		)

		if _, err := h.alertService.CreateAlertAndNotify(ctx, cmd); err != nil {
			h.logger.Printf("inactivity alert error: %v", err)
			continue
		}

		h.logger.Printf(
			"inactivity alert persisted user_id=%s device_id=%s rule_id=%s",
			userID,
			deviceID,
			rule.InactivityRuleID,
		)
	}

	return nil
}

func parseIntegrationEvent(topic string, payload []byte) (integrationEvent, error) {
	body := map[string]any{}
	if err := json.Unmarshal(payload, &body); err != nil {
		return integrationEvent{}, err
	}

	values := make(map[string]any, len(body))
	for key, value := range body {
		if key == "data" {
			continue
		}
		values[key] = value
	}

	if data, ok := body["data"].(map[string]any); ok {
		for key, value := range data {
			values[key] = value
		}
	}

	return integrationEvent{
		Topic:     topic,
		EventType: firstNonEmptyString(values, []string{"eventType", "event_type", "event"}, strings.TrimSpace(topic)),
		Values:    values,
	}, nil
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
			registerMetric(metrics, name, value)
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
	for _, candidate := range metricAliases(metric) {
		if value, ok := metrics[candidate]; ok {
			return value, true
		}
	}
	return 0, false
}

func buildThresholdMessage(metric string, value float64, thresholdValue float64, operator string, topic string, eventType string) string {
	return fmt.Sprintf(
		"Metric %s value %s %s %s from event %s on topic %s",
		metric,
		formatFloat(value),
		operator,
		formatFloat(thresholdValue),
		eventType,
		topic,
	)
}

func buildGenericMessage(topic string, eventType string, values map[string]any, fallbackTitle string) string {
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

	return fmt.Sprintf("%s received from event %s on topic %s", fallbackTitle, eventType, topic)
}

func effectiveEventType(event integrationEvent) string {
	return strings.ToLower(strings.TrimSpace(event.EventType))
}

func (h *IntegrationEventHandler) logDiscardedEvent(event integrationEvent, reason string) {
	h.logger.Printf(
		"event discarded topic=%s eventType=%s user_id=%s device_id=%s reason=%s",
		event.Topic,
		effectiveEventType(event),
		firstNonEmptyString(event.Values, []string{"user_id", "userId", "owner_id", "customer_id", "account_id"}, "n/a"),
		firstNonEmptyString(event.Values, []string{"device_id", "deviceId"}, "n/a"),
		reason,
	)
}

func matchesAnyEventType(eventType string, supported ...string) bool {
	normalized := strings.ToLower(strings.TrimSpace(eventType))
	for _, candidate := range supported {
		if normalized == strings.ToLower(strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
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

func registerMetric(metrics map[string]float64, name string, value float64) {
	for _, alias := range metricAliases(name) {
		metrics[alias] = value
	}
}

func metricAliases(metric string) []string {
	normalized := strings.ToLower(strings.TrimSpace(metric))
	switch normalized {
	case "power", "power_watts", "power_w":
		return []string{"power", "power_watts", "power_w"}
	case "energy", "energy_kwh", "kwh":
		return []string{"energy", "energy_kwh", "kwh"}
	case "cost", "estimated_cost":
		return []string{"cost", "estimated_cost"}
	default:
		if normalized == "" {
			return nil
		}
		return []string{normalized}
	}
}

func sortedMetricNames(metrics map[string]float64) []string {
	names := make([]string, 0, len(metrics))
	for name := range metrics {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
