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

type ThresholdCommandService struct {
	repo   outboundservices.AlertThresholdRepository
	logger *log.Logger
}

func NewThresholdCommandService(repo outboundservices.AlertThresholdRepository, logger *log.Logger) *ThresholdCommandService {
	return &ThresholdCommandService{repo: repo, logger: logger}
}

func (s *ThresholdCommandService) CreateThreshold(ctx context.Context, cmd commands.CreateThresholdCommand) (*entities.AlertThreshold, error) {
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

	if err := s.repo.Create(ctx, threshold); err != nil {
		return nil, err
	}

	return threshold, nil
}
