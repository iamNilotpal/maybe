package maybe_test

import (
	"testing"

	"github.com/iamNilotpal/maybe"
)

// TestOptionBasic verifies the correctness of core Option type functionality.
// It checks creation of Some/None values, access patterns, and panic behavior.
func TestOptionBasic(t *testing.T) {
	t.Run("Some/None", func(t *testing.T) {
		// Create a Some containing the integer 42.
		some := maybe.Some(42)
		// A valid Some should report true for IsSome and false for IsNone.
		if !some.IsSome() || some.IsNone() {
			t.Error("Some should be present and not None")
		}

		// Create a None instance of type int.
		none := maybe.None[int]()
		// A valid None should report false for IsSome and true for IsNone.
		if none.IsSome() || !none.IsNone() {
			t.Error("None should be absent and not Some")
		}
	})

	t.Run("Value access", func(t *testing.T) {
		// Accessing the value from a Some should return the value and true.
		opt := maybe.Some("test")
		if val, ok := opt.Value(); !ok || val != "test" {
			t.Error("Expected value 'test' and ok == true from Some")
		}

		// Accessing the value from a None should return zero-value and false.
		none := maybe.None[string]()
		if val, ok := none.Value(); ok || val != "" {
			t.Error("Expected zero value and ok == false from None")
		}
	})

	t.Run("Unwrap panic", func(t *testing.T) {
		// Unwrap on a None should panic. This test captures the panic.
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when unwrapping a None")
			}
		}()
		maybe.None[int]().Unwrap()
	})
}

// TestOptionJSON ensures Option[T] can be correctly serialized and deserialized using JSON.
// It verifies expected output for Some and None during marshaling and unmarshaling.
func TestOptionJSON(t *testing.T) {
	t.Run("Marshal", func(t *testing.T) {
		// Marshal a Some containing 42; expected output: "42"
		some := maybe.Some(42)
		data, err := some.MarshalJSON()
		if err != nil || string(data) != "42" {
			t.Errorf("Marshal of Some failed: got %s, err: %v", data, err)
		}

		// Marshal a None; expected output: "null"
		none := maybe.None[int]()
		data, err = none.MarshalJSON()
		if err != nil || string(data) != "null" {
			t.Errorf("Marshal of None failed: got %s, err: %v", data, err)
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		// Unmarshal JSON value "42" into an Option[int]; should result in Some(42)
		var some maybe.Option[int]
		err := some.UnmarshalJSON([]byte("42"))
		if err != nil || !some.IsSome() || some.Unwrap() != 42 {
			t.Errorf("Unmarshal JSON '42' failed: got %v, err: %v", some, err)
		}

		// Unmarshal JSON null into Option[int]; should result in None
		var none maybe.Option[int]
		err = none.UnmarshalJSON([]byte("null"))
		if err != nil || none.IsSome() {
			t.Errorf("Unmarshal JSON 'null' failed: got %v, err: %v", none, err)
		}
	})
}

// BenchmarkOptionValueAccess benchmarks how quickly values can be accessed from a Some[T].
func BenchmarkOptionValueAccess(b *testing.B) {
	opt := maybe.Some(42)
	for b.Loop() {
		opt.Value()
	}
}
