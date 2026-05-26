package repositories

import (
    "context"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "microservice-alert-service/alert/domain/model/entities"
    "microservice-alert-service/alert/infrastructure/persistence/gorm/model"
)

type InactivityRuleRepository struct {
    db *gorm.DB
}

func NewInactivityRuleRepository(db *gorm.DB) *InactivityRuleRepository {
    return &InactivityRuleRepository{db: db}
}

func (r *InactivityRuleRepository) Create(ctx context.Context, rule *entities.InactivityRule) error {
    record := toInactivityRuleModel(rule)
    return r.db.WithContext(ctx).Create(&record).Error
}

func (r *InactivityRuleRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.InactivityRule, error) {
    var records []model.InactivityRuleModel
    if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
        return nil, err
    }

    result := make([]entities.InactivityRule, 0, len(records))
    for _, record := range records {
        result = append(result, toInactivityRuleEntity(record))
    }

    return result, nil
}

func (r *InactivityRuleRepository) ListActiveByUserDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) ([]entities.InactivityRule, error) {
    var records []model.InactivityRuleModel
    if err := r.db.WithContext(ctx).
        Where("user_id = ? AND device_id = ? AND active = true", userID, deviceID).
        Find(&records).Error; err != nil {
        return nil, err
    }

    result := make([]entities.InactivityRule, 0, len(records))
    for _, record := range records {
        result = append(result, toInactivityRuleEntity(record))
    }

    return result, nil
}
