package commands

import (
    "time"

    "github.com/google/uuid"
)

type UpdateAlertStatusCommand struct {
    AlertID    uuid.UUID
    Status     string
    ResolvedAt *time.Time
}
