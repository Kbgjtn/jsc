package jcs

import "errors"

var (
	// ErrUnsupportedType is returned when the encoder encounters a value
	// of an unsupported type. The encoder only supports specific types
	// like integers, strings, maps, and slices. Custom structs or other
	// complex types may trigger this error.
	ErrUnsupportedType = errors.New("jcs: value has unsupported type")

	// ErrNaN is returned when the encoder encounters a NaN (Not a Number)
	// value. RFC 8785 disallows NaN values in the canonical JSON format,
	// so this error is triggered when an attempt is made to encode a NaN.
	ErrNaN = errors.New("jcs: cannot c14n NaN")

	// ErrInf is returned when the encoder encounters an Inf (infinity) value.
	// RFC 8785 also disallows Inf values (both +Inf and -Inf) in the canonical
	// JSON format, which leads to this error when such values are encountered.
	ErrInf = errors.New("jcs: cannot c14n Inf")

	// ErrInvalidUTF8 is returned when the encoder encounters a string containing
	// invalid UTF-8 byte sequences. JCS requires that all strings be valid UTF-8,
	// so this error is triggered if a string contains non-UTF-8 characters.
	ErrInvalidUTF8 = errors.New("jcs: value has invalid utf8 character")

	// ErrNumberOOR is returned when a number exceeds the valid range for precise
	// representation in the IEEE-754 double precision format (±2^53).
	// Numbers larger than ±2^53 cannot be exactly represented, and JCS requires
	// exact round-trip encoding. This error occurs when such a number is encountered.
	ErrNumberOOR = errors.New("jcs: value number out of range (v ± 2^53)")
)
