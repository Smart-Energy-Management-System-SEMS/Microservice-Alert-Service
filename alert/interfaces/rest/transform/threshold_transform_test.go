package transform

import (
	"testing"

	"microservice-alert-service/alert/interfaces/rest/resources"
)

func TestToCreateThresholdCommandDefaultsActiveToTrue(t *testing.T) {
	cmd, err := ToCreateThresholdCommand(resources.CreateThresholdRequest{
		UserID:         "f3d928b8-a5e1-4d7c-9833-a01dd5b1d258",
		DeviceID:       "2d152551-10d9-410e-a5d3-06168999735d",
		ThresholdName:  "High power",
		Metric:         "power",
		Operator:       ">=",
		ThresholdValue: 100,
	})
	if err != nil {
		t.Fatalf("ToCreateThresholdCommand() error = %v", err)
	}

	if !cmd.Active {
		t.Fatalf("expected Active default to true")
	}
}
