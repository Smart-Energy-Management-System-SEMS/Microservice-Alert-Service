package services

import (
	"fmt"

	"microservice-alert-service/alert/domain/model/valueobjects"
)

func EvaluateThreshold(operator string, value float64, threshold float64) (bool, error) {
	switch valueobjects.Operator(operator) {
	case valueobjects.OperatorGreaterThan:
		return value > threshold, nil
	case valueobjects.OperatorGreaterThanOrEqual:
		return value >= threshold, nil
	case valueobjects.OperatorLessThan:
		return value < threshold, nil
	case valueobjects.OperatorLessThanOrEqual:
		return value <= threshold, nil
	case valueobjects.OperatorEqual:
		return value == threshold, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}
