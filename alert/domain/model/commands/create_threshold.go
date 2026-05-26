package commands

import "github.com/google/uuid"

type CreateThresholdCommand struct {
    UserID         uuid.UUID
    DeviceID       uuid.UUID
    ThresholdName  string
    Metric         string
    Operator       string
    ThresholdValue float64
    Active         bool
}
