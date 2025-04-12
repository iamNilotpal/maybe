# maybe - A Go Package for Optionals, Nullables, and Nil-Safe Data Handling

`maybe` is a Go package providing type-safe optional values and functional
programming utilities using Go generics. It helps eliminate nil pointer panics
and provides a more expressive way to handle optional values in Go. The package
brings modern constructs like `Option[T]` and `Nullable[T]` to Go with strong
database and JSON integration, along with functional utilities for cleaner data
transformation.

## Table of Contents

- [Features](#features)
- [Why Use maybe?](#why-use-maybe)
- [Installation](#installation)
- [Core Types](#core-types)
  - [Option](#option)
  - [Nullable](#nullable)
- [Usage Examples](#usage-examples)
  - [Basic Operations](#basic-operations)
  - [Working with Functions](#working-with-functions)
  - [Database Integration](#database-integration)
  - [JSON Handling](#json-handling)
  - [Functional Programming](#functional-programming)
  - [Advanced Usage](#advanced-usage)
  - [Utility Functions](#utility-functions)
- [API Reference](#api-reference)
  - [Option Methods](#option-methods)
  - [Nullable Methods](#nullable-methods)
  - [Utility Functions](#utility-functions-reference)
- [License](#license)

## Features

- `Option[T]`: Type-safe optional values (Some/None) for any type.
- `Nullable[T]`: Optional values specifically designed for database and JSON
  null values.
- Functional programming utilities (Map, Filter, Reduce, etc.).
- JSON marshaling/unmarshaling support.
- SQL database integration with `database/sql` compatibility.
- Zero dependencies beyond the standard library.
- Fully compatible with Go generics (Go 1.18+).

## Why Use `maybe`?

- **Type Safety**: Avoid nil pointer panics and make optional values explicit.
- **Expressiveness**: Clearer intent than nil pointers or pointer returns.
- **Composability**: Functional utilities for working with collections.
- **Interoperability**: Easy conversion between options, nullables, and
  pointers.
- **Maintainability**: More readable code with explicit handling of missing
  values.
- **Database Integration**: First-class support for handling SQL NULL values.
- **JSON Compatibility**: Seamless handling of missing or null JSON fields.

## Installation

```bash
go get github.com/iamNilotpal/maybe
```

## Core Types

### Option

`Option[T]` represents an optional value: either `Some(value)` or `None`. It
provides a type-safe alternative to nil pointers and helps avoid nil pointer
panics. The zero value of `Option` is `None` (no value present).

Key operations:

- `Some(value)`: Create an Option with a value.
- `None[T]()`: Create an Option with no value.
- `FromPtr(ptr)`: Convert a pointer to an Option.
- `IsSome()`: Check if value is present.
- `IsNone()`: Check if value is absent.
- `Value()`: Get the value and a success boolean.
- `ValueOr(default)`: Get the value or a default.
- `Unwrap()`: Get the value or panic.
- `Ptr()`: Convert to a pointer (nil if None).

### Nullable

`Nullable[T]` represents a value that might be null, designed specifically for
handling null values in databases and JSON. Unlike `Option[T]`, which is for
general-purpose optional values, `Nullable[T]` is optimized for scenarios
involving external systems that use null values.

Key operations:

- `NullableOf(value)`: Create a valid Nullable.
- `Null[T]()`: Create a null Nullable.
- `NullableFromPtr(ptr)`: Create a Nullable from a pointer.
- `IsNull()`: Check if null.
- `IsValid()`: Check if not null.
- `Extract()`: Get the value and validity.
- `ExtractOr(default)`: Get the value or a default.
- `ToPtr()`: Convert to pointer (nil if null).
- `ToOption()`: Convert to Option type.

## Usage Examples

### Basic Operations

```go
package main

import (
	"fmt"

	"github.com/iamNilotpal/maybe"
)

func main() {
	// Working with Option
	name := maybe.Some("Nilotpal")
	emptyName := maybe.None[string]()

	// Checking value presence
	fmt.Println("Has name:", name.IsSome())            // true
	fmt.Println("Has empty name:", emptyName.IsNone()) // true

	// Safe access patterns
	if value, ok := name.Value(); ok {
		fmt.Println("Name:", value) // "Nilotpal"
	}

	// Default values
	fmt.Println("Name or default:", name.ValueOr("Anonymous"))            // "Nilotpal"
	fmt.Println("Empty name or default:", emptyName.ValueOr("Anonymous")) // "Anonymous"

	// Unwrap (safe only when you're certain the value exists)
	fmt.Println("Unwrapped name:", name.Unwrap()) // "Nilotpal"
	// emptyName.Unwrap() would panic with ErrMissingValue

	// Get or set a value
	emptyName.Set("Bob")
	fmt.Println("Name after set:", emptyName.ValueOr("")) // "Bob"

	emptyName.Unset()
	fmt.Println("Is name none after unset:", emptyName.IsNone()) // true

	// Convert to/from pointers
	var _ *string = name.Ptr()           // Pointer to "Nilotpal"
	var nilPtr *string = emptyName.Ptr() // nil pointer

	someStr := "Hello"
	_ = maybe.FromPtr(&someStr) // Some("Hello")
	_ = maybe.FromPtr(nilPtr)   // None

	// Working with Nullable (for database/JSON)
	userID := maybe.NullableOf(123)
	noID := maybe.Null[int]()

	fmt.Println("Has ID:", userID.IsValid()) // true
	fmt.Println("No ID:", noID.IsNull())     // true

	// Extract value from Nullable
	if val, ok := userID.Extract(); ok {
		fmt.Println("User ID:", val) // 123
	}

	// Default values with Nullable
	fmt.Println("ID or default:", userID.ExtractOr(0))    // 123
	fmt.Println("No ID or default:", noID.ExtractOr(999)) // 999

	// Convert between Option and Nullable
	_ = userID.ToOption() // Some(123)

	// Create Nullable from pointer
	ptrID := userID.ToPtr()          // Pointer to 123
	_ = maybe.NullableFromPtr(ptrID) // Valid Nullable with 123

	// Using with zero values
	zeroInt := maybe.NullableOf(0)                     // Valid Nullable containing 0
	fmt.Println("Is zero int null?", zeroInt.IsNull()) // false

	// Equality check
	anotherZero := maybe.NullableOf(0)
	fmt.Println("Equal zero values:", zeroInt.Equals(anotherZero))   // true
	fmt.Println("Equal to different value:", zeroInt.Equals(userID)) // false
	fmt.Println("Both null equal:", noID.Equals(maybe.Null[int]()))  // true
}
```

### Working with Functions

```go
package main

import (
    "fmt"
    "strconv"

    "github.com/iamNilotpal/maybe"
)

// A function that may fail
func divide(a, b int) maybe.Option[int] {
    if b == 0 {
        return maybe.None[int]()
    }
    return maybe.Some(a / b)
}

// A function that transforms an option
func double(o maybe.Option[int]) maybe.Option[int] {
    if val, ok := o.Value(); ok {
        return maybe.Some(val * 2)
    }
    return maybe.None[int]()
}

func main() {
    // Chaining operations
    result := divide(10, 2).AndThen(double)
    fmt.Println("10/2*2 =", result.ValueOr(0))  // 10

    // Error propagation
    result = divide(10, 0).AndThen(double)
    fmt.Println("10/0*2 =", result.ValueOr(0))  // 0 (using default)

    // Using default values in chains
    result = maybe.None[int]().AndThenOr(10, double)
    fmt.Println("Default*2 =", result.ValueOr(0))  // 20 (None replaced with 10, then doubled)

    // Parse strings to numbers safely
    parseNumber := func(s string) maybe.Option[int] {
        n, err := strconv.Atoi(s)
        if err != nil {
            return maybe.None[int]()
        }
        return maybe.Some(n)
    }

    fmt.Println("Parsed '42':", parseNumber("42").ValueOr(0))    // 42
    fmt.Println("Parsed 'abc':", parseNumber("abc").ValueOr(0))  // 0 (default)

    // Using TryMap to transform a slice of strings to numbers
    validInputs := []string{"1", "2", "3"}
    validNumbers := maybe.TryMap(validInputs, parseNumber)

    if numbers, ok := validNumbers.Value(); ok {
        fmt.Println("Valid numbers:", numbers)  // [1 2 3]
    }

    invalidInputs := []string{"1", "two", "3"}
    invalidResult := maybe.TryMap(invalidInputs, parseNumber)

    if invalidResult.IsNone() {
        fmt.Println("Failed to parse all inputs")  // This will print
    }
}
```

### Database Integration

```go
package main

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/iamNilotpal/maybe"
)

type User struct {
    ID        int
    Name      string
    Email     maybe.Nullable[string]    // Can be NULL in database
    LastLogin maybe.Nullable[time.Time] // Can be NULL in database
    Age       maybe.Nullable[int]       // Can be NULL in database
}

func main() {
    // Open database connection
    db, err := sql.Open("mysql", "user:password@/dbName")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Query a user - directly using the Scan method
    user, err := queryUser(db, 1)
    if err != nil {
        panic(err)
    }

    // Use the values
    fmt.Println("User:", user.Name)

    if email, ok := user.Email.Extract(); ok {
        fmt.Println("Email:", email)
    } else {
        fmt.Println("No email provided")
    }

    if login, ok := user.LastLogin.Extract(); ok {
        fmt.Println("Last login:", login.Format(time.RFC1123))
    } else {
        fmt.Println("Never logged in")
    }

    // Update user - directly using the Value method
    err = updateUser(db, user)
    if err != nil {
        panic(err)
    }
}

func queryUser(db *sql.DB, id int) (User, error) {
    var user User

    query := `SELECT id, name, email, last_login, age FROM users WHERE id = ?`
    err := db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Name,
        &user.Email,  // Scan directly into Nullable
        &user.LastLogin,
        &user.Age,
    )

    if err != nil {
        return User{}, err
    }

    return user, nil
}

func updateUser(db *sql.DB, user User) error {
    query := `UPDATE users SET email = ?, last_login = ?, age = ? WHERE id = ?`

    // Value() implements driver.Valuer, so we can pass the Nullable directly
    _, err := db.Exec(query,
        user.Email,     // Value() is called internally
        user.LastLogin, // Value() is called internally
        user.Age,       // Value() is called internally
        user.ID,
    )

    return err
}
```

### JSON Handling

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/iamNilotpal/maybe"
)

type Person struct {
    Name     string                 `json:"name"`
    Age      maybe.Option[int]      `json:"age,omitempty"`   // Omitted if not present
    Phone    maybe.Nullable[string] `json:"phone"`           // Explicit null if not present
    Address  maybe.Option[Address]  `json:"address,omitempty"`
}

type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

func main() {
    // Creating a person with some fields missing
    person := Person{
        Name:  "John Doe",
        Age:   maybe.Some(30),
        Phone: maybe.Null[string](),  // Explicitly null
        // Address is implicitly None
    }

    // Marshal to JSON
    data, _ := json.MarshalIndent(person, "", "  ")
    fmt.Println(string(data))
    // Output:
    // {
    //   "name": "John Doe",
    //   "age": 30,
    //   "phone": null
    // }

    // JSON with explicit null vs missing field
    jsonData := []byte(`{
        "name": "Nilotpal Deka",
        "age": null,
        "phone": "555-1234",
        "address": {
            "street": "123 Main St",
            "city": "Guwahati",
            "country": "IND"
        }
    }`)

    var anotherPerson Person
    json.Unmarshal(jsonData, &anotherPerson)

    // Age was null in JSON, so it's None in our struct
    fmt.Println("Has age:", anotherPerson.Age.IsSome())  // false

    // Phone was present, so it's a valid Nullable
    fmt.Println("Has phone:", anotherPerson.Phone.IsValid())  // true
    if phone, ok := anotherPerson.Phone.Extract(); ok {
        fmt.Println("Phone:", phone)  // "555-1234"
    }

    // Address was present as an object
    if address, ok := anotherPerson.Address.Value(); ok {
        fmt.Println("City:", address.City)  // "Guwahati"
    }

    // Error handling with MarshalJSON
    badJSON := []byte(`{"name": "Bad Data", "phone": ["invalid"]}`)
    var badPerson Person
    err := json.Unmarshal(badJSON, &badPerson)
    if err != nil {
        fmt.Println("JSON error:", err)
    }
}
```

### Functional Programming

```go
package main

import (
    "fmt"
    "strings"

    "github.com/iamNilotpal/maybe"
)

type Product struct {
    ID       int
    Name     string
    Price    float64
    Category string
    InStock  bool
}

func main() {
    products := []Product{
        {1, "Phone", 699.99, "Electronics", true},
        {2, "Laptop", 1299.99, "Electronics", false},
        {3, "Headphones", 149.99, "Audio", true},
        {4, "Monitor", 349.99, "Electronics", true},
        {5, "Speaker", 89.99, "Audio", false},
    }

    // MapSlice: Transform each product to its name
    productNames := maybe.MapSlice(products, func(p Product) string {
        return p.Name
    })
    fmt.Println("Product names:", productNames)
    // Output: [Phone Laptop Headphones Monitor Speaker]

    // FilterSlice: Get products in stock
    inStockProducts := maybe.FilterSlice(products, func(p Product) bool {
        return p.InStock
    })
    fmt.Println("In-stock products count:", len(inStockProducts))  // 3

    // ReduceSlice: Calculate total price of all products
    totalPrice := maybe.ReduceSlice(products, 0.0, func(total float64, p Product) float64 {
        return total + p.Price
    })
    fmt.Println("Total price:", totalPrice)  // 2589.95

    // ForEachSlice: Print each product name with category
    fmt.Println("Products with categories:")
    maybe.ForEachSlice(products, func(p Product) {
        fmt.Printf("- %s (%s)\n", p.Name, p.Category)
    })
    // Output:
    // - Phone (Electronics)
    // - Laptop (Electronics)
    // - Headphones (Audio)
    // - Monitor (Electronics)
    // - Speaker (Audio)

    // Combine operations: Find names of in-stock electronics products in uppercase
    electronicsInStock := maybe.FilterSlice(products, func(p Product) bool {
        return p.Category == "Electronics" && p.InStock
    })

    electronicsNames := maybe.MapSlice(electronicsInStock, func(p Product) string {
        return strings.ToUpper(p.Name)
    })

    fmt.Println("In-stock electronics:", electronicsNames) // [PHONE MONITOR]

    // Working with Options
    optionalProducts := []maybe.Option[Product]{
        maybe.Some(products[0]),
        maybe.None[Product](),
        maybe.Some(products[2]),
    }

    // CollectOptions: Get all valid products, or None if any are missing
    allProducts := maybe.CollectOptions(optionalProducts)
    fmt.Println("All products collected:", allProducts.IsSome()) // false

    // FilterSomeOptions: Get only the valid products, ignoring None
    validProducts := maybe.FilterSomeOptions(optionalProducts)
    fmt.Println("Valid products count:", len(validProducts)) // 2

    // PartitionOptions: Split into valid products and indices of missing products
    validProds, missingIndices := maybe.PartitionOptions(optionalProducts)
    fmt.Println("Valid products:", len(validProds))          // 2
    fmt.Println("Missing product indices:", missingIndices)  // [1]
}
```

### Advanced Usage

```go
package main

import (
    "fmt"
    "strconv"

    "github.com/iamNilotpal/maybe"
)

// Parse a string to an int, returning an Option
func parseToInt(s string) maybe.Option[int] {
    n, err := strconv.Atoi(s)
    if err != nil {
        return maybe.None[int]()
    }
    return maybe.Some(n)
}

// Check if an integer is positive
func ensurePositive(n maybe.Option[int]) maybe.Option[int] {
    if val, ok := n.Value(); ok && val > 0 {
        return maybe.Some(val)
    }
    return maybe.None[int]()
}

// Double a number
func doubleNumber(n maybe.Option[int]) maybe.Option[int] {
    if val, ok := n.Value(); ok {
        return maybe.Some(val * 2)
    }
    return maybe.None[int]()
}

func main() {
    // Chaining operations with AndThen
    result := parseToInt("42").
        AndThen(ensurePositive).
        AndThen(doubleNumber)

    fmt.Println("Result:", result.ValueOr(0)) // 84

    // Chain breaks on first None
    badResult := parseToInt("-10").
        AndThen(ensurePositive). // This returns None
        AndThen(doubleNumber)    // This is not called

    fmt.Println("Bad result:", badResult.ValueOr(0)) // 0

    // Using AndThenOr with default value
    withDefault := maybe.None[int]().
        AndThenOr(5, doubleNumber)

    fmt.Println("With default:", withDefault.ValueOr(0)) // 10

    // TryMap for performing operations that might fail
    inputs := []string{"1", "2", "3"}
    mapped := maybe.TryMap(inputs, parseToInt)

    if numbers, ok := mapped.Value(); ok {
        sum := maybe.ReduceSlice(numbers, 0, func(acc, n int) int {
            return acc + n
        })
        fmt.Println("Sum of valid inputs:", sum) // 6
    }

    // Failing case
    badInputs := []string{"1", "two", "3"}
    badMapped := maybe.TryMap(badInputs, parseToInt)

    fmt.Println("All inputs valid:", badMapped.IsSome()) // false
}
```

### Utility Functions

```go
package main

import (
    "fmt"

    "github.com/iamNilotpal/maybe"
)

type Person struct {
  Name string
  Age  int
}

func main() {
    // IsZero: Check for zero values of different types
    fmt.Println("0 is zero:", maybe.IsZero(0))           // true
    fmt.Println("Empty string is zero:", maybe.IsZero("")) // true
    fmt.Println("42 is zero:", maybe.IsZero(42))         // false
    fmt.Println("Hello is zero:", maybe.IsZero("Hello")) // false

    // Zero value for custom struct
    fmt.Println("Empty struct is zero:", maybe.IsZero(Person{})) // true
    fmt.Println("Non-empty struct is zero:", maybe.IsZero(Person{Name: "Nilotpal"})) // false

    // IsNil: Check for nil values
    var nilMap map[string]int
    var nilSlice []int
    var nilChan chan int
    var nilPtr *int

    fmt.Println("nil map is nil:", maybe.IsNil(nilMap))     // true
    fmt.Println("nil slice is nil:", maybe.IsNil(nilSlice)) // true
    fmt.Println("nil chan is nil:", maybe.IsNil(nilChan))   // true
    fmt.Println("nil ptr is nil:", maybe.IsNil(nilPtr))     // true

    // Non-nil values
    slice := []int{1, 2, 3}
    fmt.Println("Non-empty slice is nil:", maybe.IsNil(slice)) // false

    // Primitive types are never nil
    fmt.Println("Integer is nil:", maybe.IsNil(42)) // false

    // FirstNonZero: Find first non-zero value in a sequence
    value, found := maybe.FirstNonZero(0, "", "fallback", "ignored")
    fmt.Println("First non-zero:", value, found) // "fallback", true

    // All values are zero
    noValue, found := maybe.FirstNonZero(0, "", false, 0.0)
    fmt.Println("No valid value found:", noValue, found) // false, false

    // Complex example with multiple possible sources
    name := ""              // Primary source (empty)
    defaultName := "Guest"  // Default value
    fallbackName := "User"  // Another fallback

    username, _ := maybe.FirstNonZero(name, defaultName, fallbackName)
    fmt.Println("Username:", username) // "Guest"
}
```

## API Reference

### Option Methods

- **`Some[T](value T) Option[T]`**: Creates a new Option containing the provided
  value.
- **`None[T]() Option[T]`**: Creates a new Option with no value.
- **`FromPtr[T](ptr *T) Option[T]`**: Converts a pointer to an Option. Returns
  `Some(value)` if pointer is non-nil, or `None` if pointer is nil.
- **`Set(v T)`**: Updates the Option to contain the provided value. Changes None
  to Some(value) or updates an existing Some value.
- **`Unset()`**: Clears the Option, changing it to None. The contained value is
  set to the zero value.
- **`IsSome() bool`**: Returns true if the Option contains a value.
- **`IsNone() bool`**: Returns true if the Option does not contain a value.
- **`Value() (T, bool)`**: Returns the contained value and a boolean indicating
  if the value is present.
- **`ValueOr(defaultValue T) T`**: Returns the contained value if present,
  otherwise returns the provided default value.
- **`Ptr() *T`**: Converts the Option to a pointer. Returns a pointer to the
  value if Some, or nil if None.
- **`Unwrap() T`**: Returns the contained value if present. Panics with
  ErrMissingValue if the Option is None.
- **`UnwrapOr(defaultValue T) T`**: Returns the contained value if present,
  otherwise returns the provided default value.
- **`AndThen(fn func(Option[T]) Option[T]) Option[T]`**: Chains Option
  operations, executing the provided function only if the Option is Some.
- **`AndThenOr(defaultValue T, fn func(Option[T]) Option[T]) Option[T]`**:
  Chains Option operations but uses the provided default value if the Option is
  None.
- **`MarshalJSON() ([]byte, error)`**: Marshals the Option to JSON. Returns an
  error if the Option is None.
- **`UnmarshalJSON(data []byte) error`**: Unmarshal JSON data into the Option,
  setting it to None if the JSON value is null.

### Nullable Methods

- **`NullableOf[T](value T) Nullable[T]`**: Creates a valid Nullable with the
  provided value.
- **`Null[T]() Nullable[T]`**: Creates an invalid (null) Nullable.
- **`NullableFromPtr[T](ptr *T) Nullable[T]`**: Creates a Nullable from a
  pointer. If the pointer is nil, returns an invalid Nullable.
- **`IsNull() bool`**: Returns true if this represents a null value.
- **`IsValid() bool`**: Returns true if this represents a non-null value.
- **`Extract() (T, bool)`**: Returns the contained value and a boolean
  indicating if the value is valid.
- **`ExtractOr(defaultVal T) T`**: Returns the value if valid, otherwise returns
  the default.
- **`ToPtr() *T`**: Converts to a pointer, which will be nil if the value is
  null.
- **`ToOption() Option[T]`**: Converts Nullable to an Option type.
- **`Equals(other Nullable[T]) bool`**: Compares two Nullable values for
  equality.
- **`MarshalJSON() ([]byte, error)`**: Implements the json.Marshaler interface.
  An invalid Nullable will be marshaled as null.
- **`UnmarshalJSON(data []byte) error`**: Implements the json.Unmarshaler
  interface. A null JSON value will be unmarshaled as an invalid Nullable.
- **`Value() (driver.Value, error)`**: Implements the driver.Valuer interface
  for database operations.
- **`Scan(value any) error`**: Implements the sql.Scanner interface for database
  operations.

### Utility Functions Reference

- **`IsZero[T comparable](v T) bool`**: Tests if a value is the zero value for
  its type.
- **`IsNil(i any) bool`**: Tests if a value is nil. Works with pointer types and
  handles the case where the interface itself is nil.
- **`FirstNonZero[T comparable](vals ...T) (T, bool)`**: Returns the first
  non-zero value from the provided values.
- **`MapSlice[T, U any](input []T, mapFn func(T) U) []U`**: Applies a function
  to each element in a slice and returns a new slice with the results.
- **`FilterSlice[T any](input []T, predicate func(T) bool) []T`**: Returns a new
  slice containing only the elements for which the predicate returns true.
- **`ReduceSlice[T, R any](input []T, initial R, reducer func(R, T) R) R`**:
  Applies a function to each element in a slice, accumulating a result.
- **`ForEachSlice[T any](input []T, fn func(T))`**: Executes a function for each
  element in a slice.
- **`CollectOptions[T any](options []Option[T]) Option[[]T]`**: Transforms a
  slice of Options into an Option containing a slice of all Some values.
- **`FilterSomeOptions[T any](options []Option[T]) []T`**: Returns a slice
  containing only the values from non-empty Options.
- **`PartitionOptions[T any](options []Option[T]) (values []T, noneIndices []int)`**:
  Separates a slice of Options into values from Some options and indices of None
  options.
- **`TryMap[T, U any](input []T, fn func(T) Option[U]) Option[[]U]`**: Applies a
  function that might fail to each element in a slice.

## License

MIT License - see [LICENSE](LICENSE) for details.
