package commands

import (
	"time"

	"github.com/google/uuid"
)

type CreateAlertCommand struct {
	UserID           uuid.UUID
	DeviceID         uuid.UUID
	ThresholdID      *uuid.UUID
	InactivityRuleID *uuid.UUID
	AlertType        string
	Title            string
	Message          string
	Severity         string
	Status           string
	TriggeredAt      time.Time
}
