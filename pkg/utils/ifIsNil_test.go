package utils

import (
	"testing"
)

func TestIfIsNil(t *testing.T) {
	t.Parallel()

	// Test with string type
	t.Run("String type with nil value", func(t *testing.T) {
		var nilStr *string
		defaultStr := "default"
		result := IfIsNil(nilStr, defaultStr)
		if result != defaultStr {
			t.Errorf("Expected default value %q, got %q", defaultStr, result)
		}
	})

	t.Run("String type with non-nil value", func(t *testing.T) {
		str := "actual"
		defaultStr := "default"
		result := IfIsNil(&str, defaultStr)
		if result != str {
			t.Errorf("Expected actual value %q, got %q", str, result)
		}
	})

	// Test with int type
	t.Run("Int type with nil value", func(t *testing.T) {
		var nilInt *int
		defaultInt := 42
		result := IfIsNil(nilInt, defaultInt)
		if result != defaultInt {
			t.Errorf("Expected default value %d, got %d", defaultInt, result)
		}
	})

	t.Run("Int type with non-nil value", func(t *testing.T) {
		num := 100
		defaultInt := 42
		result := IfIsNil(&num, defaultInt)
		if result != num {
			t.Errorf("Expected actual value %d, got %d", num, result)
		}
	})

	// Test with bool type
	t.Run("Bool type with nil value", func(t *testing.T) {
		var nilBool *bool
		defaultBool := true
		result := IfIsNil(nilBool, defaultBool)
		if result != defaultBool {
			t.Errorf("Expected default value %t, got %t", defaultBool, result)
		}
	})

	t.Run("Bool type with non-nil value", func(t *testing.T) {
		boolVal := false
		defaultBool := true
		result := IfIsNil(&boolVal, defaultBool)
		if result != boolVal {
			t.Errorf("Expected actual value %t, got %t", boolVal, result)
		}
	})
}