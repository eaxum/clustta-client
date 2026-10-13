package settings

import "testing"

func TestIsValidInterfaceTextSize(t *testing.T) {
	tests := []struct {
		name  string
		size  int
		valid bool
	}{
		{name: "minimum", size: minimumInterfaceTextSize, valid: true},
		{name: "default", size: defaultInterfaceTextSize, valid: true},
		{name: "maximum", size: maximumInterfaceTextSize, valid: true},
		{name: "below minimum", size: minimumInterfaceTextSize - 1, valid: false},
		{name: "above maximum", size: maximumInterfaceTextSize + 1, valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidInterfaceTextSize(test.size); got != test.valid {
				t.Fatalf("isValidInterfaceTextSize(%d) = %t, want %t", test.size, got, test.valid)
			}
		})
	}
}
