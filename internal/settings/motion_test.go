package settings

import "testing"

func TestIsValidMotionPreference(t *testing.T) {
	tests := []struct {
		name       string
		preference string
		valid      bool
	}{
		{name: "system", preference: motionPreferenceSystem, valid: true},
		{name: "on", preference: motionPreferenceOn, valid: true},
		{name: "off", preference: motionPreferenceOff, valid: true},
		{name: "empty", preference: "", valid: false},
		{name: "unknown", preference: "sometimes", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidMotionPreference(test.preference); got != test.valid {
				t.Fatalf("isValidMotionPreference(%q) = %t, want %t", test.preference, got, test.valid)
			}
		})
	}
}
