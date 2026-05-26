package entities

import (
    "time"

    "github.com/google/uuid"
)

type Alert struct {
    AlertID          uuid.UUID
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
    ResolvedAt       *time.Time
}
