package maybe

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

// Nullable[T] represents a value that might be null.
// Unlike Option, Nullable is specifically designed for handling
// null values in external systems like databases and JSON APIs.
// This is particularly useful for database operations where NULL
// values are common and need distinct handling.
type Nullable[T any] struct {
	value T
	valid bool
}

// NullableOf creates a valid Nullable with the provided value.
func NullableOf[T any](value T) Nullable[T] {
	return Nullable[T]{value: value, valid: true}
}

// Null creates an invalid (null) Nullable.
func Null[T any]() Nullable[T] {
	var zero T
	return Nullable[T]{value: zero, valid: false}
}

// NullableFromPtr creates a Nullable from a pointer.
// If the pointer is nil, returns an invalid Nullable.
func NullableFromPtr[T any](ptr *T) Nullable[T] {
	if IsNil(ptr) {
		return Null[T]()
	}
	return NullableOf(*ptr)
}

// IsNull returns true if this represents a null value.
func (n Nullable[T]) IsNull() bool {
	return !n.valid
}

// IsValid returns true if this represents a non-null value.
func (n Nullable[T]) IsValid() bool {
	return n.valid
}

// Value returns the contained value and a boolean indicating if the value is valid.
// If the Nullable is null, returns the zero value of T and false.
func (n Nullable[T]) Extract() (T, bool) {
	return n.value, n.valid
}

// ExtractOr returns the value if valid, otherwise returns the default.
func (n Nullable[T]) ExtractOr(defaultVal T) T {
	if n.valid {
		return n.value
	}
	return defaultVal
}

// ToPtr converts to a pointer, which will be nil if the value is null.
func (n Nullable[T]) ToPtr() *T {
	if !n.valid {
		return nil
	}
	return &n.value
}

// ToOption converts Nullable to an Option type.
// This allows for interoperability between the two optional value representations.
func (n Nullable[T]) ToOption() Option[T] {
	if !n.valid {
		return None[T]()
	}
	return Some(n.value)
}

// Equals compares two Nullable values for equality.
// Two Nullable values are equal if:
//  1. Both are null, or
//  2. Both are valid and contain equal values.
func (n Nullable[T]) Equals(other Nullable[T]) bool {
	if n.valid != other.valid {
		return false
	}

	if !n.valid {
		return true
	}

	// Check if T is comparable.
	if tType := reflect.TypeOf(n.value); !tType.Comparable() {
		return false
	}

	return reflect.DeepEqual(n.value, other.value)
}

// MarshalJSON implements the json.Marshaler interface.
// An invalid Nullable will be marshaled as null.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// A null JSON value will be unmarshaled as an invalid Nullable.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.valid = false
		var zero T
		n.value = zero
		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	n.value = v
	n.valid = true
	return nil
}

// Value implements driver.Valuer to convert Go types to database-compatible values.
// This method is used when inserting parameters into SQL statements.
func (n Nullable[T]) Value() (driver.Value, error) {
	// Handle NULL case first: if not valid, return nil to represent SQL NULL
	if !n.valid {
		return nil, nil
	}

	// First check if the value implements driver.Valuer itself (e.g., custom types)
	// This allows custom types to handle their own SQL conversion logic
	if valuer, ok := any(n.value).(driver.Valuer); ok {
		return valuer.Value()
	}

	// Fast path for common types that don't need conversion
	// These types are directly supported by SQL drivers
	switch v := any(n.value).(type) {
	case int64, float64, bool, []byte, string, time.Time:
		return v, nil
	}

	// Reflection-based handling for other numeric types
	rv := reflect.ValueOf(n.value)
	switch rv.Kind() {
	// Convert signed integers to int64 (widely supported by SQL drivers)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return rv.Int(), nil

	// Handle unsigned integers with overflow checking
	// SQL typically doesn't support unsigned integers directly, so we convert to int64
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		unsignedVal := rv.Uint()
		// Prevent overflow when converting large uint64 to int64
		if unsignedVal > math.MaxInt64 {
			return nil, fmt.Errorf("unsigned integer overflow: %v exceeds int64 maximum", unsignedVal)
		}
		return int64(unsignedVal), nil

	// Convert float32 to float64 (SQL standard floating point type)
	case reflect.Float32:
		return float64(rv.Float()), nil
	}

	// Fallback error for unsupported types
	return nil, fmt.Errorf("unsupported database type: %T", n.value)
}

// Scan implements sql.Scanner to convert database values to Go types.
// This method is used when reading rows from database results.
func (n *Nullable[T]) Scan(value any) error {
	// Handle NULL case first: set valid=false and reset value
	if IsNil(value) {
		n.valid = false
		n.value = *new(T) // Zero value for type T
		return nil
	}

	// First check if the destination type implements sql.Scanner
	// This allows custom types to handle their own scanning logic
	if scanner, ok := any(&n.value).(sql.Scanner); ok {
		err := scanner.Scan(value)
		if err != nil {
			return err
		}
		n.valid = true
		return nil
	}

	// Fast path: direct type match between source and destination
	if val, ok := value.(T); ok {
		n.value = val
		n.valid = true
		return nil
	}

	// Reflection-based conversion system for type mismatches
	destType := reflect.TypeOf(n.value)
	sourceVal := reflect.ValueOf(value)

	// Numeric type handling (integers and floats)
	switch destType.Kind() {
	// Integer family (both signed and unsigned)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		var intVal int64
		var err error

		// Handle different source types that might contain integer values
		switch v := value.(type) {
		case int64: // Direct from SQL driver (most common case)
			intVal = v
		case []byte: // Textual representation from some drivers
			intVal, err = strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse integer from text: %w", err)
			}
		default:
			return fmt.Errorf("unsupported source type %T for integer conversion", value)
		}

		// Overflow checking for signed destination types
		if destType.Kind() >= reflect.Int && destType.Kind() <= reflect.Int64 {
			minVal := reflect.Zero(destType).Int()
			maxVal := reflect.New(destType).Elem().Int()
			if intVal < minVal || intVal > maxVal {
				return fmt.Errorf("value %d overflows destination type %s", intVal, destType)
			}
		} else { // Unsigned destination types
			// Prevent negative values for unsigned types
			if intVal < 0 {
				return fmt.Errorf("negative value %d for unsigned type %s", intVal, destType)
			}
			// Check for uint64 overflow
			maxUint := reflect.New(destType).Elem().OverflowUint(math.MaxUint64)
			if maxUint {
				return fmt.Errorf("value %d overflows unsigned type %s", intVal, destType)
			}
		}

		// Perform the conversion after validation
		n.value = sourceVal.Convert(destType).Interface().(T)

	// String type handling
	case reflect.String:
		switch v := value.(type) {
		case []byte: // Common case for TEXT/VARCHAR columns
			n.value = reflect.ValueOf(string(v)).Convert(destType).Interface().(T)
		default:
			return fmt.Errorf("unsupported source type %T for string conversion", value)
		}

	// Floating point type handling
	case reflect.Float32, reflect.Float64:
		var floatVal float64

		switch v := value.(type) {
		case float64: // Direct from SQL driver
			floatVal = v
		case []byte: // Textual representation from some drivers
			val, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return fmt.Errorf("failed to parse float from text: %w", err)
			}
			floatVal = val
		default:
			return fmt.Errorf("unsupported source type %T for float conversion", value)
		}

		// Float32 range checking
		if destType.Kind() == reflect.Float32 {
			if floatVal < -math.MaxFloat32 || floatVal > math.MaxFloat32 {
				return fmt.Errorf("value %f overflows float32 range", floatVal)
			}
		}

		n.value = reflect.ValueOf(floatVal).Convert(destType).Interface().(T)

	default:
		return fmt.Errorf("cannot scan %T into Nullable[%T]", value, n.value)
	}

	n.valid = true
	return nil
}
