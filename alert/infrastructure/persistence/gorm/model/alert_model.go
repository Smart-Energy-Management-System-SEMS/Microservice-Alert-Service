package model

import (
	"time"

	"github.com/google/uuid"
)

type AlertModel struct {
	AlertID          uuid.UUID  `gorm:"column:alert_id;type:uuid;primaryKey"`
	UserID           uuid.UUID  `gorm:"column:user_id;type:uuid;index"`
	DeviceID         uuid.UUID  `gorm:"column:device_id;type:uuid;index"`
	ThresholdID      *uuid.UUID `gorm:"column:threshold_id;type:uuid"`
	InactivityRuleID *uuid.UUID `gorm:"column:inactivity_rule_id;type:uuid"`
	AlertType        string     `gorm:"column:alert_type;size:50"`
	Title            string     `gorm:"column:title;size:150"`
	Message          string     `gorm:"column:message;type:text"`
	Severity         string     `gorm:"column:severity;size:30"`
	Status           string     `gorm:"column:status;size:30"`
	TriggeredAt      time.Time  `gorm:"column:triggered_at"`
	ResolvedAt       *time.Time `gorm:"column:resolved_at"`
}

func (AlertModel) TableName() string {
	return "alerts"
}
