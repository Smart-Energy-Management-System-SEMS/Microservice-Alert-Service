package entities

import (
	"time"

	"github.com/google/uuid"
)

type NotificationLog struct {
	NotificationID uuid.UUID
	AlertID        uuid.UUID
	Channel        string
	Recipient      string
	Status         string
	SentAt         *time.Time
	ErrorMessage   *string
	CreatedAt      time.Time
}
