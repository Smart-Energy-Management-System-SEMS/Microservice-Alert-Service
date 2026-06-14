// repositories.go — Repository ports (interfaces) for every aggregate the
// Application layer persists or reads. The domain depends on these contracts,
// not on a concrete database; the infrastructure layer provides the
// implementations (Dependency Inversion).

package outboundservices

import (
	"context"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/domain/model/entities"
)

// AlertRepository persists and retrieves Alert aggregates.
type AlertRepository interface {
	Create(ctx context.Context, alert *entities.Alert) error
	UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, resolvedAt *time.Time) error
	GetByID(ctx context.Context, alertID uuid.UUID) (*entities.Alert, error)
	ListAll(ctx context.Context) ([]entities.Alert, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Alert, error)
}

// AlertEventPublisher emits integration events after an alert is created.
type AlertEventPublisher interface {
	PublishJSON(ctx context.Context, key string, payload any) error
	Topic() string
}

// AlertThresholdRepository persists thresholds and lists the active ones used
// during event evaluation.
type AlertThresholdRepository interface {
	Create(ctx context.Context, threshold *entities.AlertThreshold) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.AlertThreshold, error)
	ListActiveByUserDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) ([]entities.AlertThreshold, error)
}

// InactivityRuleRepository persists inactivity rules and lists the active ones.
type InactivityRuleRepository interface {
	Create(ctx context.Context, rule *entities.InactivityRule) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.InactivityRule, error)
	ListActiveByUserDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) ([]entities.InactivityRule, error)
}

// NotificationPreferenceRepository persists and lists user channel preferences.
type NotificationPreferenceRepository interface {
	Create(ctx context.Context, preference *entities.NotificationPreference) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.NotificationPreference, error)
}

// NotificationLogRepository persists the audit log of notification attempts.
type NotificationLogRepository interface {
	Create(ctx context.Context, logEntry *entities.NotificationLog) error
}

// DeviceActivityRepository tracks the last time each device reported activity,
// used to evaluate inactivity rules.
type DeviceActivityRepository interface {
	GetLastActivity(ctx context.Context, deviceID uuid.UUID) (time.Time, bool, error)
	SaveLastActivity(ctx context.Context, deviceID uuid.UUID, at time.Time) error
}
