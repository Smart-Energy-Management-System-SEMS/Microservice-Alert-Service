package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type DeviceActivityRepository struct {
	mu        sync.RWMutex
	lastEvent map[uuid.UUID]time.Time
}

func NewDeviceActivityRepository() *DeviceActivityRepository {
	return &DeviceActivityRepository{lastEvent: make(map[uuid.UUID]time.Time)}
}

func (r *DeviceActivityRepository) GetLastActivity(ctx context.Context, deviceID uuid.UUID) (time.Time, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.lastEvent[deviceID]
	return value, ok, nil
}

func (r *DeviceActivityRepository) SaveLastActivity(ctx context.Context, deviceID uuid.UUID, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastEvent[deviceID] = at
	return nil
}
