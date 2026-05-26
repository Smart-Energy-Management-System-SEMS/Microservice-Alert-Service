package entities

import (
	"time"

	"github.com/google/uuid"
)

type InactivityRule struct {
	InactivityRuleID   uuid.UUID
	UserID             uuid.UUID
	DeviceID           uuid.UUID
	RuleName           string
	MaxInactiveMinutes int
	Active             bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
