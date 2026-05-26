package transform

import (
	"microservice-alert-service/alert/domain/model/commands"
	"microservice-alert-service/alert/domain/model/entities"
	"microservice-alert-service/alert/interfaces/rest/resources"
)

func ToCreateNotificationPreferenceCommand(req resources.CreateNotificationPreferenceRequest) (commands.CreateNotificationPreferenceCommand, error) {
	userID, err := parseUUID(req.UserID)
	if err != nil {
		return commands.CreateNotificationPreferenceCommand{}, err
	}

	quietStart, err := parseTimeOfDay(req.QuietHoursStart)
	if err != nil {
		return commands.CreateNotificationPreferenceCommand{}, err
	}

	quietEnd, err := parseTimeOfDay(req.QuietHoursEnd)
	if err != nil {
		return commands.CreateNotificationPreferenceCommand{}, err
	}

	return commands.CreateNotificationPreferenceCommand{
		UserID:          userID,
		Channel:         req.Channel,
		Enabled:         req.Enabled,
		MinSeverity:     req.MinSeverity,
		QuietHoursStart: quietStart,
		QuietHoursEnd:   quietEnd,
	}, nil
}

func ToNotificationPreferenceResponse(entity entities.NotificationPreference) resources.NotificationPreferenceResponse {
	return resources.NotificationPreferenceResponse{
		PreferenceID:    entity.PreferenceID.String(),
		UserID:          entity.UserID.String(),
		Channel:         entity.Channel,
		Enabled:         entity.Enabled,
		MinSeverity:     entity.MinSeverity,
		QuietHoursStart: entity.QuietHoursStart,
		QuietHoursEnd:   entity.QuietHoursEnd,
	}
}
