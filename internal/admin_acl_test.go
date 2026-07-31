package internal

import "testing"

func TestAdminACL(t *testing.T) {
	prefixes, err := parseAdminACLs("192.0.2.0/24, 10.0.0.5/32")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		source string
		want   bool
	}{
		{"192.0.2.42", true}, {"10.0.0.5", true}, {"10.0.0.6", false},
	} {
		if got := adminSourceAllowed(tt.source, prefixes); got != tt.want {
			t.Errorf("source %s: allowed %v, want %v", tt.source, got, tt.want)
		}
	}
}
