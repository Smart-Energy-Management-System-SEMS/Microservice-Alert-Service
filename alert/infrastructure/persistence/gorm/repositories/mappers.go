package repositories

import (
	"microservice-alert-service/alert/domain/model/entities"
	"microservice-alert-service/alert/infrastructure/persistence/gorm/model"
)

func toAlertModel(entity *entities.Alert) model.AlertModel {
	return model.AlertModel{
		AlertID:          entity.AlertID,
		UserID:           entity.UserID,
		DeviceID:         entity.DeviceID,
		ThresholdID:      entity.ThresholdID,
		InactivityRuleID: entity.InactivityRuleID,
		AlertType:        entity.AlertType,
		Title:            entity.Title,
		Message:          entity.Message,
		Severity:         entity.Severity,
		Status:           entity.Status,
		TriggeredAt:      entity.TriggeredAt,
		ResolvedAt:       entity.ResolvedAt,
	}
}

func toAlertEntity(modelData model.AlertModel) entities.Alert {
	return entities.Alert{
		AlertID:          modelData.AlertID,
		UserID:           modelData.UserID,
		DeviceID:         modelData.DeviceID,
		ThresholdID:      modelData.ThresholdID,
		InactivityRuleID: modelData.InactivityRuleID,
		AlertType:        modelData.AlertType,
		Title:            modelData.Title,
		Message:          modelData.Message,
		Severity:         modelData.Severity,
		Status:           modelData.Status,
		TriggeredAt:      modelData.TriggeredAt,
		ResolvedAt:       modelData.ResolvedAt,
	}
}

func toThresholdModel(entity *entities.AlertThreshold) model.AlertThresholdModel {
	return model.AlertThresholdModel{
		ThresholdID:    entity.ThresholdID,
		UserID:         entity.UserID,
		DeviceID:       entity.DeviceID,
		ThresholdName:  entity.ThresholdName,
		Metric:         entity.Metric,
		Operator:       entity.Operator,
		ThresholdValue: entity.ThresholdValue,
		Active:         entity.Active,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}
}

func toThresholdEntity(modelData model.AlertThresholdModel) entities.AlertThreshold {
	return entities.AlertThreshold{
		ThresholdID:    modelData.ThresholdID,
		UserID:         modelData.UserID,
		DeviceID:       modelData.DeviceID,
		ThresholdName:  modelData.ThresholdName,
		Metric:         modelData.Metric,
		Operator:       modelData.Operator,
		ThresholdValue: modelData.ThresholdValue,
		Active:         modelData.Active,
		CreatedAt:      modelData.CreatedAt,
		UpdatedAt:      modelData.UpdatedAt,
	}
}

func toInactivityRuleModel(entity *entities.InactivityRule) model.InactivityRuleModel {
	return model.InactivityRuleModel{
		InactivityRuleID:   entity.InactivityRuleID,
		UserID:             entity.UserID,
		DeviceID:           entity.DeviceID,
		RuleName:           entity.RuleName,
		MaxInactiveMinutes: entity.MaxInactiveMinutes,
		Active:             entity.Active,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
	}
}

func toInactivityRuleEntity(modelData model.InactivityRuleModel) entities.InactivityRule {
	return entities.InactivityRule{
		InactivityRuleID:   modelData.InactivityRuleID,
		UserID:             modelData.UserID,
		DeviceID:           modelData.DeviceID,
		RuleName:           modelData.RuleName,
		MaxInactiveMinutes: modelData.MaxInactiveMinutes,
		Active:             modelData.Active,
		CreatedAt:          modelData.CreatedAt,
		UpdatedAt:          modelData.UpdatedAt,
	}
}

func toNotificationPreferenceModel(entity *entities.NotificationPreference) model.NotificationPreferenceModel {
	return model.NotificationPreferenceModel{
		PreferenceID:    entity.PreferenceID,
		UserID:          entity.UserID,
		Channel:         entity.Channel,
		Enabled:         entity.Enabled,
		MinSeverity:     entity.MinSeverity,
		QuietHoursStart: entity.QuietHoursStart,
		QuietHoursEnd:   entity.QuietHoursEnd,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}
}

func toNotificationPreferenceEntity(modelData model.NotificationPreferenceModel) entities.NotificationPreference {
	return entities.NotificationPreference{
		PreferenceID:    modelData.PreferenceID,
		UserID:          modelData.UserID,
		Channel:         modelData.Channel,
		Enabled:         modelData.Enabled,
		MinSeverity:     modelData.MinSeverity,
		QuietHoursStart: modelData.QuietHoursStart,
		QuietHoursEnd:   modelData.QuietHoursEnd,
		CreatedAt:       modelData.CreatedAt,
		UpdatedAt:       modelData.UpdatedAt,
	}
}

func toNotificationLogModel(entity *entities.NotificationLog) model.NotificationLogModel {
	return model.NotificationLogModel{
		NotificationID: entity.NotificationID,
		AlertID:        entity.AlertID,
		Channel:        entity.Channel,
		Recipient:      entity.Recipient,
		Status:         entity.Status,
		SentAt:         entity.SentAt,
		ErrorMessage:   entity.ErrorMessage,
		CreatedAt:      entity.CreatedAt,
	}
}
