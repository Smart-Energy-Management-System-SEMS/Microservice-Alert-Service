// create_threshold.go — CQRS command to create an alert threshold.

package commands

import "github.com/google/uuid"

// CreateThresholdCommand carries the data needed to define a threshold that
// raises an alert when a device metric satisfies the operator against the
// configured value (e.g. value > 100).
type CreateThresholdCommand struct {
	UserID         uuid.UUID // owner of the threshold
	DeviceID       uuid.UUID // device the threshold watches
	ThresholdName  string    // human-readable threshold name
	Metric         string    // metric being monitored (e.g. "power")
	Operator       string    // comparison operator (e.g. ">", ">=", "<")
	ThresholdValue float64   // value to compare the metric against
	Active         bool      // whether the threshold is currently enabled
}
