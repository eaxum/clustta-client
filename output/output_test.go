package output

import (
	"encoding/json"
	"testing"
)

func TestProgressReportOptionalDownloadFields(t *testing.T) {
	for _, operation := range []string{"", "project-download"} {
		t.Run(operation, func(t *testing.T) {
			report := ProgressReport{Title: "Transfer", Message: "Receiving", Current: 1, Total: 1}
			if operation != "" {
				report.Operation = operation
				report.Phase = "receiving"
			}
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"title", "message", "percentage", "current", "total", "extra_message", "operation_type"} {
				if _, exists := payload[key]; !exists {
					t.Errorf("missing legacy field %s", key)
				}
			}
			for key, expected := range map[string]string{"operation": report.Operation, "phase": report.Phase} {
				value, exists := payload[key]
				if expected == "" && exists {
					t.Errorf("legacy event includes %s", key)
				}
				if expected != "" && value != expected {
					t.Errorf("%s = %v, want %s", key, value, expected)
				}
			}
		})
	}
}
