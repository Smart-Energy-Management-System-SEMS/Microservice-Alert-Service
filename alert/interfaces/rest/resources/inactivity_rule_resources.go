package resources

type CreateInactivityRuleRequest struct {
	UserID             string `json:"user_id" binding:"required"`
	DeviceID           string `json:"device_id" binding:"required"`
	RuleName           string `json:"rule_name" binding:"required"`
	MaxInactiveMinutes int    `json:"max_inactive_minutes" binding:"required"`
	Active             bool   `json:"active"`
}

type InactivityRuleResponse struct {
	InactivityRuleID   string `json:"inactivity_rule_id"`
	UserID             string `json:"user_id"`
	DeviceID           string `json:"device_id"`
	RuleName           string `json:"rule_name"`
	MaxInactiveMinutes int    `json:"max_inactive_minutes"`
	Active             bool   `json:"active"`
}
