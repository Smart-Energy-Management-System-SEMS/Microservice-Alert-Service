package resources

import "time"

type CreateNotificationPreferenceRequest struct {
    UserID          string `json:"user_id" binding:"required"`
    Channel         string `json:"channel" binding:"required"`
    Enabled         bool   `json:"enabled"`
    MinSeverity     string `json:"min_severity"`
    QuietHoursStart string `json:"quiet_hours_start"`
    QuietHoursEnd   string `json:"quiet_hours_end"`
}

type NotificationPreferenceResponse struct {
    PreferenceID    string     `json:"preference_id"`
    UserID          string     `json:"user_id"`
    Channel         string     `json:"channel"`
    Enabled         bool       `json:"enabled"`
    MinSeverity     string     `json:"min_severity"`
    QuietHoursStart *time.Time `json:"quiet_hours_start"`
    QuietHoursEnd   *time.Time `json:"quiet_hours_end"`
}
