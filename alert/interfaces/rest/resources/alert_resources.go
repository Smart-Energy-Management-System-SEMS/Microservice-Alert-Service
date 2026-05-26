package resources

import "time"

type CreateAlertRequest struct {
	UserID           string     `json:"user_id" binding:"required"`
	DeviceID         string     `json:"device_id" binding:"required"`
	ThresholdID      *string    `json:"threshold_id"`
	InactivityRuleID *string    `json:"inactivity_rule_id"`
	AlertType        string     `json:"alert_type" binding:"required"`
	Title            string     `json:"title" binding:"required"`
	Message          string     `json:"message" binding:"required"`
	Severity         string     `json:"severity" binding:"required"`
	Status           string     `json:"status" binding:"required"`
	TriggeredAt      *time.Time `json:"triggered_at"`
}

type UpdateAlertStatusRequest struct {
	Status     string     `json:"status" binding:"required"`
	ResolvedAt *time.Time `json:"resolved_at"`
}

type AlertResponse struct {
	AlertID          string     `json:"alert_id"`
	UserID           string     `json:"user_id"`
	DeviceID         string     `json:"device_id"`
	ThresholdID      *string    `json:"threshold_id"`
	InactivityRuleID *string    `json:"inactivity_rule_id"`
	AlertType        string     `json:"alert_type"`
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	Severity         string     `json:"severity"`
	Status           string     `json:"status"`
	TriggeredAt      time.Time  `json:"triggered_at"`
	ResolvedAt       *time.Time `json:"resolved_at"`
}
