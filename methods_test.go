package maybe_test

import (
	"testing"
	"time"

	"github.com/iamNilotpal/maybe"
)

// TestIsZero verifies the behavior of maybe.IsZero for various types.
// It checks both primitive and complex types to ensure correct zero-value detection.
func TestIsZero(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{"int zero", 0, true},                  // Zero value for int
		{"int non-zero", 42, false},            // Non-zero int
		{"string empty", "", true},             // Empty string should be zero
		{"string non-empty", "hello", false},   // Non-empty string
		{"struct zero", time.Time{}, true},     // Zero value for time.Time
		{"struct non-zero", time.Now(), false}, // Current time is non-zero
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Type switch to call maybe.IsZero with the correct generic type
			switch v := tt.input.(type) {
			case int:
				got := maybe.IsZero(v)
				if got != tt.want {
					t.Errorf("IsZero(%v) = %v, want %v", v, got, tt.want)
				}
			case string:
				got := maybe.IsZero(v)
				if got != tt.want {
					t.Errorf("IsZero(%v) = %v, want %v", v, got, tt.want)
				}
			case time.Time:
				got := maybe.IsZero(v)
				if got != tt.want {
					t.Errorf("IsZero(%v) = %v, want %v", v, got, tt.want)
				}
			}
		})
	}
}

// TestFirstNonZero checks maybe.FirstNonZero which returns the first non-zero value
// among a variadic list of values and whether such a value was found.
func TestFirstNonZero(t *testing.T) {
	t.Run("int values", func(t *testing.T) {
		// Only the third value is non-zero, should return 1
		got, ok := maybe.FirstNonZero(0, 0, 1, 42)
		if !ok || got != 1 {
			t.Errorf("FirstNonZero() = %v, %v; want 1, true", got, ok)
		}
	})

	t.Run("float values", func(t *testing.T) {
		// Third float is the first non-zero
		got, ok := maybe.FirstNonZero(0.0, 0.0, 3.14)
		if !ok || got != 3.14 {
			t.Errorf("FirstNonZero() = %v, %v; want 3.14, true", got, ok)
		}
	})

	t.Run("string values", func(t *testing.T) {
		// Third string is non-empty
		got, ok := maybe.FirstNonZero("", "", "fallback", "other")
		if !ok || got != "fallback" {
			t.Errorf("FirstNonZero() = %v, %v; want 'fallback', true", got, ok)
		}

		// All are empty strings; should return false
		got2, ok2 := maybe.FirstNonZero("", "", "")
		if ok2 {
			t.Errorf("FirstNonZero() = %v, %v; want zero value and false", got2, ok2)
		}
	})

	t.Run("struct values", func(t *testing.T) {
		type Person struct{ Name string }
		zero := Person{}
		nonZero := Person{Name: "John"}

		// nonZero should be picked as the first valid
		got, ok := maybe.FirstNonZero(zero, nonZero, Person{Name: "Doe"})
		if !ok || got != nonZero {
			t.Errorf("FirstNonZero() = %v, %v; want %v, true", got, ok, nonZero)
		}

		// Only zero-value structs provided; expect false
		got2, ok2 := maybe.FirstNonZero(zero, zero)
		if ok2 {
			t.Errorf("FirstNonZero() = %v, %v; want zero value and false", got2, ok2)
		}
	})
}

// TestMapSlice ensures that maybe.MapSlice applies the transformation function
// to each element in a slice and returns a new transformed slice.
func TestMapSlice(t *testing.T) {
	input := []int{1, 2, 3}
	double := func(n int) int { return n * 2 }

	// Should return [2, 4, 6]
	got := maybe.MapSlice(input, double)

	if len(got) != 3 || got[0] != 2 || got[1] != 4 || got[2] != 6 {
		t.Errorf("MapSlice() = %v, want [2 4 6]", got)
	}
}

// BenchmarkMapSlice measures the performance of the MapSlice function
// with a large input to evaluate allocation and transformation speed.
func BenchmarkMapSlice(b *testing.B) {
	input := make([]int, 1000) // Slice of 1000 zeros
	mapFn := func(n int) int { return n * 2 }

	// Run the benchmark loop
	for b.Loop() {
		maybe.MapSlice(input, mapFn)
	}
}
