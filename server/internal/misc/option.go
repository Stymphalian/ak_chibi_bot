package misc

import "encoding/json"

// Option holds an optional value of type T.
type Option[T any] struct {
	has   bool
	value T
}

// NewOption creates a new Option instance.
//
// If the value is nil, the Option will be None.
func NewOption[T any](value T) Option[T] {
	return Option[T]{value: value, has: true}
}

func EmptyOption[T any]() Option[T] {
	return Option[T]{has: false}
}

// IsNone checks if the Option is None.
func (o Option[T]) IsNone() bool {
	return !o.has
}

// IsSome checks if the Option is Some.
func (o Option[T]) IsSome() bool {
	return o.has
}

// Unwrap returns the value of the Option.
//
// If the Option is None, Unwrap panics.
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("called `Unwrap` on a `None` value")
	}
	return o.value
}

// UnwrapOr returns the value of the Option.
//
// If the Option is None, UnwrapOr returns the defaultValue.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.IsNone() {
		return defaultValue
	}
	return o.value
}

// MarshalJSON implements the json.Marshaler interface.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}
