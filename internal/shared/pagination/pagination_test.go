package pagination

import "testing"

func TestPageSize_Valid(t *testing.T) {
	tests := []struct {
		name string
		size PageSize
		want bool
	}{
		{"10 is valid", PageSize10, true},
		{"25 is valid", PageSize25, true},
		{"50 is valid", PageSize50, true},
		{"100 is valid", PageSize100, true},
		{"200 is valid", PageSize200, true},
		{"zero is invalid", PageSize(0), false},
		{"negative is invalid", PageSize(-1), false},
		{"non-allowlisted value is invalid", PageSize(75), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.size.Valid(); got != tt.want {
				t.Errorf("PageSize(%d).Valid() = %v, want %v", tt.size, got, tt.want)
			}
		})
	}
}
