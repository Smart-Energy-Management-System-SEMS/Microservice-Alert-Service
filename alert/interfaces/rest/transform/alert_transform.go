package transform

import (
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/model/entities"
	"microservice-alert-service/alert/interfaces/rest/resources"
)

func ToCreateAlertCommand(req resources.CreateAlertRequest) (commands.CreateAlertCommand, error) {
	userID, err := parseUUID(req.UserID)
	if err != nil {
		return commands.CreateAlertCommand{}, err
	}

	deviceID, err := parseUUID(req.DeviceID)
	if err != nil {
		return commands.CreateAlertCommand{}, err
	}

	thresholdID, err := parseOptionalUUID(req.ThresholdID)
	if err != nil {
		return commands.CreateAlertCommand{}, err
	}

	inactivityRuleID, err := parseOptionalUUID(req.InactivityRuleID)
	if err != nil {
		return commands.CreateAlertCommand{}, err
	}

	cmd := commands.CreateAlertCommand{
		UserID:           userID,
		DeviceID:         deviceID,
		ThresholdID:      thresholdID,
		InactivityRuleID: inactivityRuleID,
		AlertType:        req.AlertType,
		Title:            req.Title,
		Message:          req.Message,
		Severity:         req.Severity,
		Status:           req.Status,
	}

	if req.TriggeredAt != nil {
		cmd.TriggeredAt = *req.TriggeredAt
	}

	return cmd, nil
}

func ToUpdateAlertStatusCommand(alertID string, req resources.UpdateAlertStatusRequest) (commands.UpdateAlertStatusCommand, error) {
	parsedID, err := parseUUID(alertID)
	if err != nil {
		return commands.UpdateAlertStatusCommand{}, err
	}

	return commands.UpdateAlertStatusCommand{
		AlertID:    parsedID,
		Status:     req.Status,
		ResolvedAt: req.ResolvedAt,
	}, nil
}

func ToAlertResponse(entity entities.Alert) resources.AlertResponse {
	response := resources.AlertResponse{
		AlertID:     entity.AlertID.String(),
		UserID:      entity.UserID.String(),
		DeviceID:    entity.DeviceID.String(),
		AlertType:   entity.AlertType,
		Title:       entity.Title,
		Message:     entity.Message,
		Severity:    entity.Severity,
		Status:      entity.Status,
		TriggeredAt: entity.TriggeredAt,
		ResolvedAt:  entity.ResolvedAt,
	}

	if entity.ThresholdID != nil {
		value := entity.ThresholdID.String()
		response.ThresholdID = &value
	}

	if entity.InactivityRuleID != nil {
		value := entity.InactivityRuleID.String()
		response.InactivityRuleID = &value
	}

	return response
}
