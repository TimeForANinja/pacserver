package IP

import (
	"errors"
	"testing"
)

func TestIsValidCIDR(t *testing.T) {
	if !isValidCIDR(0) {
		t.Errorf("isValidCIDR(0) = false; want true")
	}

	tests := []struct {
		name string
		val  int
		want bool
	}{
		{"Lower Bound", 0, true},
		{"Random Valid Value", 10, true},
		{"Upper Bound", 32, true},
		{"Outside Lower Bound", -1, false},
		{"Outside Upper Bound", 33, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidCIDR(tt.val)
			if got != tt.want {
				t.Errorf("SlicesEqual(%v) = %v; want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestCIDRToNetmask(t *testing.T) {
	tests := []struct {
		name string
		val  int
		want uint32
	}{
		{"Lower Bound", 0, 0b00000000_00000000_00000000_00000000},
		{"Random valid Value", 18, 0b11111111_11111111_11000000_00000000},
		{"Upper Bound", 32, 0b11111111_11111111_11111111_11111111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cidrToNetmask(tt.val)
			if got != tt.want {
				t.Errorf("SlicesEqual(%v) = %v; want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestNewCIDR(t *testing.T) {
	tests := []struct {
		name  string
		param int
		value uint8
		error error
	}{
		{"Regular Lower Bound", 0, 0, nil},
		{"Regular Upper Bound", 32, 32, nil},
		{"Out of (lower) Range", -1, 0, ErrCIDROutOfRange},
		{"Out of (upper) Range", 33, 0, ErrCIDROutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotErr := NewCIDR(tt.param)
			if gotErr != nil && tt.error == nil {
				t.Errorf("NewCIDR(%v) failed unexpectedly: %s", tt.param, gotErr.Error())
			} else if tt.error != nil && !errors.Is(gotErr, tt.error) {
				t.Errorf("NewCIDR(%v) failed with err %s, want %s", tt.param, gotErr.Error(), ErrCIDROutOfRange.Error())
			} else if gotErr == nil && gotVal.Value != tt.value {
				t.Errorf("NewCIDR(%v).Value = %d, want %d", tt.param, gotVal.Value, tt.value)
			}
		})
	}
}

func TestNewCIDRFromString(t *testing.T) {
	tests := []struct {
		name  string
		param string
		value uint8
		error error
	}{
		{"Regular Lower Bound", "0", 0, nil},
		{"Regular Upper Bound", "32", 32, nil},
		{"Out of (lower) Range", "-1", 0, ErrCIDROutOfRange},
		{"Out of (upper) Range", "33", 0, ErrCIDROutOfRange},
		{"Param in scientific notation", "1e0", 0, ErrCIDRNotAnInt},
		{"Param is float", "0.3", 0, ErrCIDRNotAnInt},
		{"Param is string", "asdf", 0, ErrCIDRNotAnInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotErr := NewCIDRFromString(tt.param)
			if gotErr != nil && tt.error == nil {
				t.Errorf("NewCIDR(%v) failed unexpectedly: %s", tt.param, gotErr.Error())
			} else if tt.error != nil && !errors.Is(gotErr, tt.error) {
				t.Errorf("NewCIDR(%v) failed with err %s, want %s", tt.param, gotErr.Error(), ErrCIDROutOfRange.Error())
			} else if gotErr == nil && gotVal.Value != tt.value {
				t.Errorf("NewCIDR(%v).Value = %d, want %d", tt.param, gotVal.Value, tt.value)
			}
		})
	}
}
