// Package commands defines the CQRS command objects of the domain: immutable
// data carriers that express the intent to change state. They hold input data
// only — no business logic.
package commands

import (
	"time"

	"github.com/google/uuid"
)

// CreateAlertCommand carries the data needed to raise a new alert. The
// ThresholdID and InactivityRuleID are optional (pointers) because an alert
// may originate from either source, but not necessarily both.
type CreateAlertCommand struct {
	UserID           uuid.UUID  // owner of the alert
	DeviceID         uuid.UUID  // device the alert refers to
	ThresholdID      *uuid.UUID // source threshold, if triggered by one
	InactivityRuleID *uuid.UUID // source inactivity rule, if triggered by one
	AlertType        string     // e.g. "threshold" or "inactivity"
	Title            string     // short alert title
	Message          string     // human-readable description
	Severity         string     // low | medium | high | critical
	Status           string     // e.g. "open"
	TriggeredAt      time.Time  // when the condition was detected
}
