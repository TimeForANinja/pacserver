package utils

import (
	"reflect"
	"sort"
	"testing"
)

func TestMapToArray(t *testing.T) {
	t.Parallel()

	// Test with string keys and int values
	t.Run("String keys and int values", func(t *testing.T) {
		m := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}

		result := MapToArray(m)

		// Sort the result for deterministic comparison
		sort.Ints(result)

		expected := []int{1, 2, 3}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test with int keys and string values
	t.Run("Int keys and string values", func(t *testing.T) {
		m := map[int]string{
			1: "one",
			2: "two",
			3: "three",
		}

		result := MapToArray(m)

		// Sort the result for deterministic comparison
		sort.Strings(result)

		expected := []string{"one", "three", "two"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test with empty map
	t.Run("Empty map", func(t *testing.T) {
		m := map[string]int{}

		result := MapToArray(m)

		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v with length %d", result, len(result))
		}
	})

	// Test with struct values
	t.Run("Struct values", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		m := map[string]Person{
			"alice": {Name: "Alice", Age: 30},
			"bob":   {Name: "Bob", Age: 25},
		}

		result := MapToArray(m)

		// Since map iteration order is not guaranteed, we need to check if all expected values are in the result
		expected := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}

		if len(result) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(result))
		}

		// Check if all expected values are in the result
		for _, e := range expected {
			found := false
			for _, r := range result {
				if reflect.DeepEqual(e, r) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected value %v not found in result %v", e, result)
			}
		}
	})
}

func TestMapClone(t *testing.T) {
	t.Parallel()

	t.Run("clones map contents without sharing storage", func(t *testing.T) {
		original := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}

		clone := MapClone(original)

		if !reflect.DeepEqual(original, clone) {
			t.Fatalf("clone = %#v, want %#v", clone, original)
		}

		original["one"] = 10
		delete(original, "two")
		original["four"] = 4

		if clone["one"] != 1 {
			t.Fatalf("clone was mutated when original changed: %#v", clone)
		}
		if _, ok := clone["four"]; ok {
			t.Fatalf("clone unexpectedly gained new key: %#v", clone)
		}
		if _, ok := clone["two"]; !ok {
			t.Fatalf("clone unexpectedly lost key: %#v", clone)
		}
	})

	t.Run("returns empty map for nil input", func(t *testing.T) {
		var original map[string]int

		clone := MapClone(original)

		if clone == nil {
			t.Fatal("clone should be an allocated empty map")
		}
		if len(clone) != 0 {
			t.Fatalf("clone length = %d, want 0", len(clone))
		}
	})
}
