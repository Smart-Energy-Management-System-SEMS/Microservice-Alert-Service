package transform

import (
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/model/entities"
	"microservice-alert-service/alert/interfaces/rest/resources"
)

func ToCreateInactivityRuleCommand(req resources.CreateInactivityRuleRequest) (commands.CreateInactivityRuleCommand, error) {
	userID, err := parseUUID(req.UserID)
	if err != nil {
		return commands.CreateInactivityRuleCommand{}, err
	}

	deviceID, err := parseUUID(req.DeviceID)
	if err != nil {
		return commands.CreateInactivityRuleCommand{}, err
	}

	return commands.CreateInactivityRuleCommand{
		UserID:             userID,
		DeviceID:           deviceID,
		RuleName:           req.RuleName,
		MaxInactiveMinutes: req.MaxInactiveMinutes,
		Active:             req.Active,
	}, nil
}

func ToInactivityRuleResponse(entity entities.InactivityRule) resources.InactivityRuleResponse {
	return resources.InactivityRuleResponse{
		InactivityRuleID:   entity.InactivityRuleID.String(),
		UserID:             entity.UserID.String(),
		DeviceID:           entity.DeviceID.String(),
		RuleName:           entity.RuleName,
		MaxInactiveMinutes: entity.MaxInactiveMinutes,
		Active:             entity.Active,
	}
}
