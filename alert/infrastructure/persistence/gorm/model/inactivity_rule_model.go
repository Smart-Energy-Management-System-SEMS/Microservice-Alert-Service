package model

import (
    "time"

    "github.com/google/uuid"
)

type InactivityRuleModel struct {
    InactivityRuleID  uuid.UUID `gorm:"column:inactivity_rule_id;type:uuid;primaryKey"`
    UserID            uuid.UUID `gorm:"column:user_id;type:uuid;index"`
    DeviceID          uuid.UUID `gorm:"column:device_id;type:uuid;index"`
    RuleName          string    `gorm:"column:rule_name;size:120"`
    MaxInactiveMinutes int      `gorm:"column:max_inactive_minutes"`
    Active            bool      `gorm:"column:active"`
    CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (InactivityRuleModel) TableName() string {
    return "inactivity_rules"
}
