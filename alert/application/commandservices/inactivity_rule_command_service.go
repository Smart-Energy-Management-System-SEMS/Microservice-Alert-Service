// inactivity_rule_command_service.go — Command Service (write side, CQRS) for
// the InactivityRule aggregate: creates rules that flag devices as inactive.

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

// InactivityRuleCommandService orchestrates the creation of inactivity rules.
type InactivityRuleCommandService struct {
	repo   outboundservices.InactivityRuleRepository // persistence contract
	logger *log.Logger
}

// NewInactivityRuleCommandService wires the injected dependencies.
func NewInactivityRuleCommandService(repo outboundservices.InactivityRuleRepository, logger *log.Logger) *InactivityRuleCommandService {
	return &InactivityRuleCommandService{repo: repo, logger: logger}
}

// CreateRule builds an InactivityRule entity from the command and persists it.
func (s *InactivityRuleCommandService) CreateRule(ctx context.Context, cmd commands.CreateInactivityRuleCommand) (*entities.InactivityRule, error) {
	// Map the command to a domain entity with a fresh id and UTC timestamps.
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

	// Persist; propagate any error to the caller.
	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}
