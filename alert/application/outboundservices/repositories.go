package outboundservices

import (
    "context"
    "time"

    "github.com/google/uuid"

    "microservice-alert-service/alert/domain/model/entities"
)

type AlertRepository interface {
    Create(ctx context.Context, alert *entities.Alert) error
    UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, resolvedAt *time.Time) error
    GetByID(ctx context.Context, alertID uuid.UUID) (*entities.Alert, error)
    ListAll(ctx context.Context) ([]entities.Alert, error)
    ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Alert, error)
}

type AlertThresholdRepository interface {
    Create(ctx context.Context, threshold *entities.AlertThreshold) error
    ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.AlertThreshold, error)
    ListActiveByUserDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) ([]entities.AlertThreshold, error)
}

type InactivityRuleRepository interface {
    Create(ctx context.Context, rule *entities.InactivityRule) error
    ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.InactivityRule, error)
    ListActiveByUserDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) ([]entities.InactivityRule, error)
}

type NotificationPreferenceRepository interface {
    Create(ctx context.Context, preference *entities.NotificationPreference) error
    ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.NotificationPreference, error)
}

type NotificationLogRepository interface {
    Create(ctx context.Context, logEntry *entities.NotificationLog) error
}

type DeviceActivityRepository interface {
    GetLastActivity(ctx context.Context, deviceID uuid.UUID) (time.Time, bool, error)
    SaveLastActivity(ctx context.Context, deviceID uuid.UUID, at time.Time) error
}
