// threshold_command_service.go — Command Service (write side, CQRS) for the
// AlertThreshold aggregate: creates the rules that raise an alert when a
// device metric crosses a configured value.

package commandservices

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"microservice-alert-service/alert/application/outboundservices"
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/model/entities"
)

// ThresholdCommandService orchestrates the creation of alert thresholds.
type ThresholdCommandService struct {
	repo   outboundservices.AlertThresholdRepository // persistence contract
	logger *log.Logger
}

// NewThresholdCommandService wires the injected dependencies.
func NewThresholdCommandService(repo outboundservices.AlertThresholdRepository, logger *log.Logger) *ThresholdCommandService {
	return &ThresholdCommandService{repo: repo, logger: logger}
}

// CreateThreshold builds an AlertThreshold entity from the command and stores it.
func (s *ThresholdCommandService) CreateThreshold(ctx context.Context, cmd commands.CreateThresholdCommand) (*entities.AlertThreshold, error) {
	// Map the command to a domain entity with a fresh id and UTC timestamps.
	threshold := &entities.AlertThreshold{
		ThresholdID:    uuid.New(),
		UserID:         cmd.UserID,
		DeviceID:       cmd.DeviceID,
		ThresholdName:  cmd.ThresholdName,
		Metric:         cmd.Metric,
		Operator:       cmd.Operator,
		ThresholdValue: cmd.ThresholdValue,
		Active:         cmd.Active,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// Persist; propagate any error to the caller.
	if err := s.repo.Create(ctx, threshold); err != nil {
		return nil, err
	}

	return threshold, nil
}
