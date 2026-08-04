package currency

import "testing"

func TestCode_Valid(t *testing.T) {
	tests := []struct {
		name string
		code Code
		want bool
	}{
		{"CAD is valid", CAD, true},
		{"USD is valid", USD, true},
		{"BRL is valid", BRL, true},
		{"empty is invalid", Code(""), false},
		{"unknown code is invalid", Code("XYZ"), false},
		{"lowercase valid code is invalid", Code("usd"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.Valid(); got != tt.want {
				t.Errorf("Code(%q).Valid() = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
