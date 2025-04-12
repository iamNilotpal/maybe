package maybe_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/iamNilotpal/maybe"
)

// TestNullableDatabase ensures that Nullable[T] integrates correctly with database/sql.
// It verifies the Scan and Value methods work for basic types and time.Time.
func TestNullableDatabase(t *testing.T) {
	t.Run("Scan/Value", func(t *testing.T) {
		// Test scanning a non-null string into Nullable[string]
		var n maybe.Nullable[string]
		err := n.Scan("test")
		if err != nil || !n.IsValid() || n.ExtractOr("") != "test" {
			t.Error("Scan of valid string into Nullable[string] failed")
		}

		// Test scanning a SQL NULL into Nullable[int]
		var null maybe.Nullable[int]
		err = null.Scan(nil)
		if err != nil || null.IsValid() {
			t.Error("Scan of nil into Nullable[int] should yield invalid value")
		}

		// Test that Nullable[int] correctly implements driver.Valuer
		val, err := maybe.NullableOf(42).Value()
		if err != nil || val != int64(42) {
			t.Error("Value() should return int64(42) for Nullable[int]")
		}
	})

	t.Run("Time handling", func(t *testing.T) {
		// Test scanning a time.Time value
		now := time.Now()
		var nt maybe.Nullable[time.Time]
		err := nt.Scan(now)
		val, ok := nt.Extract()

		if err != nil || !nt.IsValid() || !ok || !val.Equal(now) {
			t.Error("Scan of time.Time into Nullable failed or mismatched")
		}
	})
}

// TestNullableJSON verifies that Nullable[T] properly marshals to and from JSON.
// It checks both null and non-null cases.
func TestNullableJSON(t *testing.T) {
	t.Run("Marshal/Unmarshal", func(t *testing.T) {
		// Marshal a valid Nullable[int] -> should produce "42"
		n := maybe.NullableOf(42)
		data, err := json.Marshal(n)
		if err != nil || string(data) != "42" {
			t.Errorf("Marshal valid failed: got %s, err: %v", data, err)
		}

		// Marshal a null Nullable[int] -> should produce "null"
		null := maybe.Null[int]()
		data, err = json.Marshal(null)
		if err != nil || string(data) != "null" {
			t.Errorf("Marshal null failed: got %s, err: %v", data, err)
		}

		// Unmarshal "3.14" into Nullable[float64]
		var unm maybe.Nullable[float64]
		err = json.Unmarshal([]byte("3.14"), &unm)
		val, ok := unm.Extract()
		if err != nil || !unm.IsValid() || !ok || val != 3.14 {
			t.Errorf("Unmarshal float64 failed: got %v, ok: %v, err: %v", val, ok, err)
		}
	})
}

// TestNullableConversion checks conversion from Nullable to Option type.
func TestNullableConversion(t *testing.T) {
	t.Run("ToOption", func(t *testing.T) {
		// Converting a null Nullable[string] should yield None
		null := maybe.Null[string]()
		opt1 := null.ToOption()
		if opt1.IsSome() {
			t.Error("Null -> Option should result in None")
		}

		// Converting a valid Nullable[int] should yield Some with same value
		valid := maybe.NullableOf(42)
		opt2 := valid.ToOption()
		if !opt2.IsSome() || opt2.Unwrap() != 42 {
			t.Error("Valid Nullable -> Option conversion failed")
		}
	})
}

// BenchmarkNullableScan measures performance of repeatedly scanning a value into Nullable[T].
func BenchmarkNullableScan(b *testing.B) {
	var n maybe.Nullable[int]
	value := 42
	for b.Loop() {
		n.Scan(value)
	}
}

// TestDriverValuerImplementation ensures Nullable[T] satisfies required interfaces for SQL drivers.
func TestDriverValuerImplementation(t *testing.T) {
	// Confirm that Nullable[T] implements driver.Valuer and sql.Scanner interfaces.
	var _ driver.Valuer = maybe.Nullable[int]{}
	var _ sql.Scanner = &maybe.Nullable[int]{}
}
