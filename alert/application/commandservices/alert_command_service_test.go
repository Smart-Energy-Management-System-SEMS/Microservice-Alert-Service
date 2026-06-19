package commandservices

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/model/entities"
)

func TestCreateAlertNormalizesOpenStatusToOpen(t *testing.T) {
	repo := &alertRepoSpy{}
	service := NewAlertCommandService(repo, nil, nil, "open", log.New(io.Discard, "", 0))

	_, err := service.CreateAlert(context.Background(), commands.CreateAlertCommand{
		UserID:      uuid.New(),
		DeviceID:    uuid.New(),
		AlertType:   "threshold",
		Title:       "High power",
		Message:     "Threshold exceeded",
		Severity:    "high",
		Status:      "open",
		TriggeredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateAlert() error = %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(repo.created))
	}

	if repo.created[0].Status != "open" {
		t.Fatalf("expected normalized status open, got %s", repo.created[0].Status)
	}
}

func TestCreateAlertNormalizesPendingStatusToOpen(t *testing.T) {
	repo := &alertRepoSpy{}
	service := NewAlertCommandService(repo, nil, nil, "pending", log.New(io.Discard, "", 0))

	_, err := service.CreateAlert(context.Background(), commands.CreateAlertCommand{
		UserID:      uuid.New(),
		DeviceID:    uuid.New(),
		AlertType:   "threshold",
		Title:       "High power",
		Message:     "Threshold exceeded",
		Severity:    "high",
		Status:      "pending",
		TriggeredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateAlert() error = %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(repo.created))
	}

	if repo.created[0].Status != "open" {
		t.Fatalf("expected normalized status open, got %s", repo.created[0].Status)
	}
}

func TestBuildAlertCreatedEventUsesStandardEnvelope(t *testing.T) {
	triggeredAt := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	alert := &entities.Alert{
		AlertID:     uuid.New(),
		UserID:      uuid.New(),
		DeviceID:    uuid.New(),
		AlertType:   "threshold",
		Title:       "High power",
		Message:     "Threshold exceeded",
		Severity:    "high",
		Status:      "open",
		TriggeredAt: triggeredAt,
	}

	event := buildAlertCreatedEvent(alert)

	if event["eventType"] != "alert.created" {
		t.Fatalf("unexpected eventType: %v", event["eventType"])
	}
	if event["eventId"] != alert.AlertID.String() {
		t.Fatalf("unexpected eventId: %v", event["eventId"])
	}
	if event["occurredAt"] != triggeredAt.Format(time.RFC3339) {
		t.Fatalf("unexpected occurredAt: %v", event["occurredAt"])
	}
	if _, exists := event["alert_id"]; exists {
		t.Fatal("did not expect alert_id duplicated at root")
	}
	if _, exists := event["event"]; exists {
		t.Fatal("did not expect legacy event field")
	}

	data, ok := event["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %T", event["data"])
	}
	if data["alert_id"] != alert.AlertID.String() {
		t.Fatalf("unexpected data.alert_id: %v", data["alert_id"])
	}
}

type alertRepoSpy struct {
	created []*entities.Alert
}

func (r *alertRepoSpy) Create(_ context.Context, alert *entities.Alert) error {
	copy := *alert
	r.created = append(r.created, &copy)
	return nil
}

func (r *alertRepoSpy) UpdateStatus(context.Context, uuid.UUID, string, *time.Time) error {
	return nil
}

func (r *alertRepoSpy) GetByID(context.Context, uuid.UUID) (*entities.Alert, error) {
	return nil, nil
}

func (r *alertRepoSpy) ListAll(context.Context) ([]entities.Alert, error) {
	return nil, nil
}

func (r *alertRepoSpy) ListByUser(context.Context, uuid.UUID) ([]entities.Alert, error) {
	return nil, nil
}
