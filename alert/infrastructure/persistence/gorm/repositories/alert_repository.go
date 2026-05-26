package repositories

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "microservice-alert-service/alert/domain/model/entities"
    "microservice-alert-service/alert/shared/domain"
    "microservice-alert-service/alert/infrastructure/persistence/gorm/model"
)

type AlertRepository struct {
    db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
    return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(ctx context.Context, alert *entities.Alert) error {
    record := toAlertModel(alert)
    return r.db.WithContext(ctx).Create(&record).Error
}

func (r *AlertRepository) UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, resolvedAt *time.Time) error {
    updates := map[string]interface{}{"status": status, "resolved_at": resolvedAt}
    return r.db.WithContext(ctx).Model(&model.AlertModel{}).Where("alert_id = ?", alertID).Updates(updates).Error
}

func (r *AlertRepository) GetByID(ctx context.Context, alertID uuid.UUID) (*entities.Alert, error) {
    var record model.AlertModel
    err := r.db.WithContext(ctx).First(&record, "alert_id = ?", alertID).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrNotFound
        }
        return nil, err
    }

    entity := toAlertEntity(record)
    return &entity, nil
}

func (r *AlertRepository) ListAll(ctx context.Context) ([]entities.Alert, error) {
    var records []model.AlertModel
    if err := r.db.WithContext(ctx).Find(&records).Error; err != nil {
        return nil, err
    }

    result := make([]entities.Alert, 0, len(records))
    for _, record := range records {
        result = append(result, toAlertEntity(record))
    }

    return result, nil
}

func (r *AlertRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.Alert, error) {
    var records []model.AlertModel
    if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
        return nil, err
    }

    result := make([]entities.Alert, 0, len(records))
    for _, record := range records {
        result = append(result, toAlertEntity(record))
    }

    return result, nil
}
