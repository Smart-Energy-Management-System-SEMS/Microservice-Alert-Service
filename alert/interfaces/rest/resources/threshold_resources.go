package resources

type CreateThresholdRequest struct {
	UserID         string  `json:"user_id" binding:"required"`
	DeviceID       string  `json:"device_id" binding:"required"`
	ThresholdName  string  `json:"threshold_name" binding:"required"`
	Metric         string  `json:"metric" binding:"required"`
	Operator       string  `json:"operator" binding:"required"`
	ThresholdValue float64 `json:"threshold_value" binding:"required"`
	Active         bool    `json:"active"`
}

type ThresholdResponse struct {
	ThresholdID    string  `json:"threshold_id"`
	UserID         string  `json:"user_id"`
	DeviceID       string  `json:"device_id"`
	ThresholdName  string  `json:"threshold_name"`
	Metric         string  `json:"metric"`
	Operator       string  `json:"operator"`
	ThresholdValue float64 `json:"threshold_value"`
	Active         bool    `json:"active"`
}
