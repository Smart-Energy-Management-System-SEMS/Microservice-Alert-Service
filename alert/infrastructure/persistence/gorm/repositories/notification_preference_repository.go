package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"microservice-alert-service/alert/domain/model/entities"
	"microservice-alert-service/alert/infrastructure/persistence/gorm/model"
)

type NotificationPreferenceRepository struct {
	db *gorm.DB
}

func NewNotificationPreferenceRepository(db *gorm.DB) *NotificationPreferenceRepository {
	return &NotificationPreferenceRepository{db: db}
}

func (r *NotificationPreferenceRepository) Create(ctx context.Context, preference *entities.NotificationPreference) error {
	record := toNotificationPreferenceModel(preference)
	return r.db.WithContext(ctx).Create(&record).Error
}

func (r *NotificationPreferenceRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.NotificationPreference, error) {
	var records []model.NotificationPreferenceModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]entities.NotificationPreference, 0, len(records))
	for _, record := range records {
		result = append(result, toNotificationPreferenceEntity(record))
	}

	return result, nil
}
