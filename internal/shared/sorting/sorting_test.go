package sorting

import "testing"

func TestDirection_Valid(t *testing.T) {
	tests := []struct {
		name string
		dir  Direction
		want bool
	}{
		{"asc is valid", Ascending, true},
		{"desc is valid", Descending, true},
		{"empty is invalid", Direction(""), false},
		{"unknown direction is invalid", Direction("sideways"), false},
		{"uppercase valid direction is invalid", Direction("ASC"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dir.Valid(); got != tt.want {
				t.Errorf("Direction(%q).Valid() = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}
