package commands

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotificationPreferenceCommand struct {
	UserID          uuid.UUID
	Channel         string
	Enabled         bool
	MinSeverity     string
	QuietHoursStart *time.Time
	QuietHoursEnd   *time.Time
}
