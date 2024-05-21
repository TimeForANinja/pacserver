package utils

import (
	"testing"
)

func TestSlicesEqual_string(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{"EqualStringSlices", []string{"a", "b", "c"}, []string{"a", "b", "c"}, true},
		{"DifferentStringSlices", []string{"a", "b", "c"}, []string{"a", "b", "d"}, false},
		{"DifferentLengthStringSlices", []string{"a", "b", "c"}, []string{"a", "b"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlicesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("SlicesEqual(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSlicesEqual_int(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		b    []int
		want bool
	}{
		{"EqualIntSlices", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"DifferentIntSlices", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"DifferentLengthIntSlices", []int{1, 2, 3}, []int{1, 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlicesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("SlicesEqual(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSlicesEqual_Nil(t *testing.T) {
	// Test case for comparing nil slices
	var a []int
	var b []int
	if !SlicesEqual(a, b) {
		t.Errorf("SlicesEqual(%v, %v) = %v; want %v", a, b, false, true)
	}
}
