package model

import (
    "time"

    "github.com/google/uuid"
)

type AlertThresholdModel struct {
    ThresholdID    uuid.UUID `gorm:"column:threshold_id;type:uuid;primaryKey"`
    UserID         uuid.UUID `gorm:"column:user_id;type:uuid;index"`
    DeviceID       uuid.UUID `gorm:"column:device_id;type:uuid;index"`
    ThresholdName  string    `gorm:"column:threshold_name;size:120"`
    Metric         string    `gorm:"column:metric;size:50"`
    Operator       string    `gorm:"column:operator;size:10"`
    ThresholdValue float64   `gorm:"column:threshold_value;type:numeric(12,4)"`
    Active         bool      `gorm:"column:active"`
    CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (AlertThresholdModel) TableName() string {
    return "alert_thresholds"
}
