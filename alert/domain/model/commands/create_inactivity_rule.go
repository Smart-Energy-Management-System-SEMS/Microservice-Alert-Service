// create_inactivity_rule.go — CQRS command to create an inactivity rule.

package commands

import "github.com/google/uuid"

// CreateInactivityRuleCommand carries the data needed to define a rule that
// flags a device as inactive after a configured number of minutes.
type CreateInactivityRuleCommand struct {
	UserID             uuid.UUID // owner of the rule
	DeviceID           uuid.UUID // device the rule watches
	RuleName           string    // human-readable rule name
	MaxInactiveMinutes int       // inactivity limit before alerting
	Active             bool       // whether the rule is currently enabled
}
