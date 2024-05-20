package IP

import (
	"testing"
)

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"Valid Full (private) IP", "10.3.3.2", true},
		{"Valid Full (public) IP", "8.8.8.8", true},
		{"Invalid Partial IP", "102.1.2", false},
		{"Invalid partial IP with tracing dot", "201.3.", false},
		{"Invalid IP with tuple out of range", "400.1.2.3", false},
		{"Invalid IP with tuple out of range", "1.2.3.400", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidIP(tt.val)
			if got != tt.want {
				t.Errorf("IsValidIP(%v) = %v; want %v", tt.val, got, tt.want)
			}
		})
	}
}
