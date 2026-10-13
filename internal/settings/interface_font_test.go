package settings

import "testing"

func TestIsValidInterfaceFont(t *testing.T) {
	tests := []struct {
		name  string
		font  string
		valid bool
	}{
		{name: "default", font: defaultInterfaceFont, valid: true},
		{name: "inter", font: interfaceFontInter, valid: true},
		{name: "sour gummy", font: interfaceFontSourGummy, valid: true},
		{name: "empty", font: "", valid: false},
		{name: "unknown", font: "unknown", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidInterfaceFont(test.font); got != test.valid {
				t.Fatalf("isValidInterfaceFont(%q) = %t, want %t", test.font, got, test.valid)
			}
		})
	}
}
