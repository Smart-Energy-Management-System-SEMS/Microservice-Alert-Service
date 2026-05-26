package entities

import (
    "time"

    "github.com/google/uuid"
)

type NotificationPreference struct {
    PreferenceID    uuid.UUID
    UserID          uuid.UUID
    Channel         string
    Enabled         bool
    MinSeverity     string
    QuietHoursStart *time.Time
    QuietHoursEnd   *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
