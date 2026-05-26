package commands

import "github.com/google/uuid"

type CreateInactivityRuleCommand struct {
	UserID             uuid.UUID
	DeviceID           uuid.UUID
	RuleName           string
	MaxInactiveMinutes int
	Active             bool
}
