package transform

import (
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/model/entities"
	"microservice-alert-service/alert/interfaces/rest/resources"
)

func ToCreateThresholdCommand(req resources.CreateThresholdRequest) (commands.CreateThresholdCommand, error) {
	userID, err := parseUUID(req.UserID)
	if err != nil {
		return commands.CreateThresholdCommand{}, err
	}

	deviceID, err := parseUUID(req.DeviceID)
	if err != nil {
		return commands.CreateThresholdCommand{}, err
	}

	return commands.CreateThresholdCommand{
		UserID:         userID,
		DeviceID:       deviceID,
		ThresholdName:  req.ThresholdName,
		Metric:         req.Metric,
		Operator:       req.Operator,
		ThresholdValue: req.ThresholdValue,
		Active:         req.Active,
	}, nil
}

func ToThresholdResponse(entity entities.AlertThreshold) resources.ThresholdResponse {
	return resources.ThresholdResponse{
		ThresholdID:    entity.ThresholdID.String(),
		UserID:         entity.UserID.String(),
		DeviceID:       entity.DeviceID.String(),
		ThresholdName:  entity.ThresholdName,
		Metric:         entity.Metric,
		Operator:       entity.Operator,
		ThresholdValue: entity.ThresholdValue,
		Active:         entity.Active,
	}
}
