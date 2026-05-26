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

type InactivityRuleCommandService struct {
	repo   outboundservices.InactivityRuleRepository
	logger *log.Logger
}

func NewInactivityRuleCommandService(repo outboundservices.InactivityRuleRepository, logger *log.Logger) *InactivityRuleCommandService {
	return &InactivityRuleCommandService{repo: repo, logger: logger}
}

func (s *InactivityRuleCommandService) CreateRule(ctx context.Context, cmd commands.CreateInactivityRuleCommand) (*entities.InactivityRule, error) {
	rule := &entities.InactivityRule{
		InactivityRuleID:   uuid.New(),
		UserID:             cmd.UserID,
		DeviceID:           cmd.DeviceID,
		RuleName:           cmd.RuleName,
		MaxInactiveMinutes: cmd.MaxInactiveMinutes,
		Active:             cmd.Active,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}
