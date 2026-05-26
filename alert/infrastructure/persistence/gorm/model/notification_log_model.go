package model

import (
    "time"

    "github.com/google/uuid"
)

type NotificationLogModel struct {
    NotificationID uuid.UUID  `gorm:"column:notification_id;type:uuid;primaryKey"`
    AlertID        uuid.UUID  `gorm:"column:alert_id;type:uuid;index"`
    Channel        string     `gorm:"column:channel;size:30"`
    Recipient      string     `gorm:"column:recipient;size:150"`
    Status         string     `gorm:"column:status;size:30"`
    SentAt         *time.Time `gorm:"column:sent_at"`
    ErrorMessage   *string    `gorm:"column:error_message;type:text"`
    CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (NotificationLogModel) TableName() string {
    return "notification_logs"
}
