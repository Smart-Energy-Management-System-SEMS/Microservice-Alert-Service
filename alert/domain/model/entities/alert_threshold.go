package entities

import (
	"time"

	"github.com/google/uuid"
)

type AlertThreshold struct {
	ThresholdID    uuid.UUID
	UserID         uuid.UUID
	DeviceID       uuid.UUID
	ThresholdName  string
	Metric         string
	Operator       string
	ThresholdValue float64
	Active         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
