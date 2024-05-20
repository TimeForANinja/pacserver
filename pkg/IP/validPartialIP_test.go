package IP

import (
	"testing"
)

func TestIsValidPartialIP(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"Valid Full (private) IP", "10.3.3.2", true},
		{"Valid Full (public) IP", "8.8.8.8", true},
		{"Valid 3/4 IP", "102.1.2", true},
		{"Valid 3/4 IP with leading zero", "01.2.1", true},
		{"Valid 2/4 IP", "201.3", true},
		{"Valid 1/4 IP", "32", true},
		{"Invalid partial IP with tracing dot", "201.3.", false},
		{"Invalid IP with tuple out of range", "1000.1.2", false},
		{"Invalid IP with tuple out of range", "1.2.300", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPartialIP(tt.val)
			if got != tt.want {
				t.Errorf("IsValidPartialIP(%v) = %v; want %v", tt.val, got, tt.want)
			}
		})
	}
}
