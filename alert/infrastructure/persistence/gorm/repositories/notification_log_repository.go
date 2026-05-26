package repositories

import (
	"context"

	"gorm.io/gorm"

	"microservice-alert-service/alert/domain/model/entities"
)

type NotificationLogRepository struct {
	db *gorm.DB
}

func NewNotificationLogRepository(db *gorm.DB) *NotificationLogRepository {
	return &NotificationLogRepository{db: db}
}

func (r *NotificationLogRepository) Create(ctx context.Context, logEntry *entities.NotificationLog) error {
	record := toNotificationLogModel(logEntry)
	return r.db.WithContext(ctx).Create(&record).Error
}
