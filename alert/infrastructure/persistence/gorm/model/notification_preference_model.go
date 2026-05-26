package model

import (
    "time"

    "github.com/google/uuid"
)

type NotificationPreferenceModel struct {
    PreferenceID    uuid.UUID  `gorm:"column:preference_id;type:uuid;primaryKey"`
    UserID          uuid.UUID  `gorm:"column:user_id;type:uuid;index"`
    Channel         string     `gorm:"column:channel;size:30"`
    Enabled         bool       `gorm:"column:enabled"`
    MinSeverity     string     `gorm:"column:min_severity;size:30"`
    QuietHoursStart *time.Time `gorm:"column:quiet_hours_start;type:time"`
    QuietHoursEnd   *time.Time `gorm:"column:quiet_hours_end;type:time"`
    CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (NotificationPreferenceModel) TableName() string {
    return "notification_preferences"
}
