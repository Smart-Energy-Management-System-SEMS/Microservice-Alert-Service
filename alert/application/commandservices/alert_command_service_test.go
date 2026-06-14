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

func TestCreateAlertNormalizesOpenStatusToPending(t *testing.T) {
	repo := &alertRepoSpy{}
	service := NewAlertCommandService(repo, nil, nil, "pending", log.New(io.Discard, "", 0))

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

	if repo.created[0].Status != "pending" {
		t.Fatalf("expected normalized status pending, got %s", repo.created[0].Status)
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
