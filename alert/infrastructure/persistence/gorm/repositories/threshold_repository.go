package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"microservice-alert-service/alert/domain/model/entities"
	"microservice-alert-service/alert/infrastructure/persistence/gorm/model"
)

type AlertThresholdRepository struct {
	db *gorm.DB
}

func NewAlertThresholdRepository(db *gorm.DB) *AlertThresholdRepository {
	return &AlertThresholdRepository{db: db}
}

func (r *AlertThresholdRepository) Create(ctx context.Context, threshold *entities.AlertThreshold) error {
	record := toThresholdModel(threshold)
	return r.db.WithContext(ctx).Create(&record).Error
}

func (r *AlertThresholdRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entities.AlertThreshold, error) {
	var records []model.AlertThresholdModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]entities.AlertThreshold, 0, len(records))
	for _, record := range records {
		result = append(result, toThresholdEntity(record))
	}

	return result, nil
}

func (r *AlertThresholdRepository) ListActiveByUserDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) ([]entities.AlertThreshold, error) {
	var records []model.AlertThresholdModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ? AND active = true", userID, deviceID).
		Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]entities.AlertThreshold, 0, len(records))
	for _, record := range records {
		result = append(result, toThresholdEntity(record))
	}

	return result, nil
}
