package eventhandlers

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/commandservices"
	"microservice-alert-service/alert/domain/model/entities"
)

func TestHandleMessageCreatesThresholdAlertFromGroupedEnergyTopic(t *testing.T) {
	userID := uuid.New()
	deviceID := uuid.New()
	thresholdID := uuid.New()

	alertRepo := &stubAlertRepository{}
	handler := NewIntegrationEventHandler(
		&stubThresholdRepository{
			thresholds: []entities.AlertThreshold{
				{
					ThresholdID:    thresholdID,
					UserID:         userID,
					DeviceID:       deviceID,
					ThresholdName:  "High power",
					Metric:         "power",
					Operator:       ">=",
					ThresholdValue: 100,
					Active:         true,
				},
			},
		},
		&stubInactivityRuleRepository{},
		&stubDeviceActivityRepository{},
		commandservices.NewAlertCommandService(alertRepo, nil, nil, "open", log.New(io.Discard, "", 0)),
		log.New(io.Discard, "", 0),
	)

	payload := []byte(`{
		"eventType":"energy.consumption.recorded",
		"occurredAt":"2026-06-13T12:00:00Z",
		"data":{
			"user_id":"` + userID.String() + `",
			"device_id":"` + deviceID.String() + `",
			"power_watts":150
		}
	}`)

	if err := handler.HandleMessage(context.Background(), "energy.events", payload); err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	if len(alertRepo.created) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alertRepo.created))
	}

	alert := alertRepo.created[0]
	if alert.UserID != userID || alert.DeviceID != deviceID {
		t.Fatalf("unexpected alert identity: user=%s device=%s", alert.UserID, alert.DeviceID)
	}
	if alert.ThresholdID == nil || *alert.ThresholdID != thresholdID {
		t.Fatalf("expected threshold id %s, got %+v", thresholdID, alert.ThresholdID)
	}
}

func TestHandleMessageRejectsMissingEventType(t *testing.T) {
	handler := NewIntegrationEventHandler(
		&stubThresholdRepository{},
		&stubInactivityRuleRepository{},
		&stubDeviceActivityRepository{},
		commandservices.NewAlertCommandService(&stubAlertRepository{}, nil, nil, "open", log.New(io.Discard, "", 0)),
		log.New(io.Discard, "", 0),
	)

	payload := []byte(`{
		"occurredAt":"2026-06-13T12:00:00Z",
		"data":{"message":"missing type"}
	}`)

	err := handler.HandleMessage(context.Background(), "energy.events", payload)
	if err == nil {
		t.Fatal("expected missing eventType error")
	}
}

type stubAlertRepository struct {
	created []*entities.Alert
}

func (r *stubAlertRepository) Create(_ context.Context, alert *entities.Alert) error {
	copy := *alert
	r.created = append(r.created, &copy)
	return nil
}

func (r *stubAlertRepository) UpdateStatus(context.Context, uuid.UUID, string, *time.Time) error {
	return nil
}

func (r *stubAlertRepository) GetByID(context.Context, uuid.UUID) (*entities.Alert, error) {
	return nil, nil
}

func (r *stubAlertRepository) ListAll(context.Context) ([]entities.Alert, error) {
	return nil, nil
}

func (r *stubAlertRepository) ListByUser(context.Context, uuid.UUID) ([]entities.Alert, error) {
	return nil, nil
}

type stubThresholdRepository struct {
	thresholds []entities.AlertThreshold
}

func (r *stubThresholdRepository) Create(context.Context, *entities.AlertThreshold) error {
	return nil
}

func (r *stubThresholdRepository) ListByUser(context.Context, uuid.UUID) ([]entities.AlertThreshold, error) {
	return nil, nil
}

func (r *stubThresholdRepository) ListActiveByUserDevice(_ context.Context, userID uuid.UUID, deviceID uuid.UUID) ([]entities.AlertThreshold, error) {
	result := make([]entities.AlertThreshold, 0, len(r.thresholds))
	for _, threshold := range r.thresholds {
		if threshold.UserID == userID && threshold.DeviceID == deviceID && threshold.Active {
			result = append(result, threshold)
		}
	}
	return result, nil
}

type stubInactivityRuleRepository struct{}

func (r *stubInactivityRuleRepository) Create(context.Context, *entities.InactivityRule) error {
	return nil
}

func (r *stubInactivityRuleRepository) ListByUser(context.Context, uuid.UUID) ([]entities.InactivityRule, error) {
	return nil, nil
}

func (r *stubInactivityRuleRepository) ListActiveByUserDevice(context.Context, uuid.UUID, uuid.UUID) ([]entities.InactivityRule, error) {
	return nil, nil
}

type stubDeviceActivityRepository struct{}

func (r *stubDeviceActivityRepository) GetLastActivity(context.Context, uuid.UUID) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

func (r *stubDeviceActivityRepository) SaveLastActivity(context.Context, uuid.UUID, time.Time) error {
	return nil
}
