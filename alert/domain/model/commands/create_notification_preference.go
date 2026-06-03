// create_notification_preference.go — CQRS command to create a user's
// notification preference.

package commands

import (
	"time"

	"github.com/google/uuid"
)

// CreateNotificationPreferenceCommand carries the data needed to define how a
// user wants to be notified. QuietHoursStart/End are optional (pointers): when
// nil, no quiet-hours window applies.
type CreateNotificationPreferenceCommand struct {
	UserID          uuid.UUID  // owner of the preference
	Channel         string     // delivery channel, e.g. "email" or "sms"
	Enabled         bool       // whether this channel is active
	MinSeverity     string     // minimum severity that triggers a notification
	QuietHoursStart *time.Time // start of the do-not-disturb window (optional)
	QuietHoursEnd   *time.Time // end of the do-not-disturb window (optional)
}
