package transform

import (
	"time"

	"github.com/google/uuid"
)

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func ParseUUID(value string) (uuid.UUID, error) {
	return parseUUID(value)
}

func parseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(*value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseTimeOfDay(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return nil, err
	}

	result := time.Date(2000, 1, 1, parsed.Hour(), parsed.Minute(), 0, 0, time.UTC)
	return &result, nil
}
