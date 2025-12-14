// Package jcs provides an implementation of the JSON Canonicalization Scheme (JCS)
// as defined in RFC 8785.
//
// JCS defines a strict, deterministic serialization of JSON data so that
// logically equivalent values always produce the same byte sequence. This
// canonical form is essential for cryptographic applications such as digital
// signatures, hashing, and secure data exchange.
//
// Features of this implementation:
//   - Canonical serialization of primitive types (null, booleans, strings, numbers)
//     following RFC 8785 rules.
//   - Enforcement of IEEE‑754 double precision constraints: integers outside
//     ±(2^53 − 1) cannot be represented exactly and return ErrNumberOOR.
//   - Canonical ordering of object keys using UTF‑16 code unit comparison,
//     ensuring correct handling of non‑BMP characters (surrogate pairs).
//   - Support for slices of common Go types (ints, uints, floats, strings, bools, any)
//     and maps with string keys.
//   - Rejection of unsupported or non‑representable types with ErrUnsupportedType.
//
// The core entry point is Append, which appends the canonical JSON representation
// of a Go value to a destination byte slice. Helper functions such as appendSlice
// and appendObject handle composite types. Errors are returned when values cannot
// be represented according to RFC 8785.
//
// This package is intended for use in contexts where canonical JSON is required
// for interoperability, compliance, or cryptographic integrity.
package jcs

import "time"

// Append function is part of the jcs package, which implements the JSON
// Canonicalization Scheme (JCS) as defined in RFC 8785. This function appends
// the canonicalized JSON representation of various Go types to a byte slice (dst).
// It returns the canonicalized representation of the provided value v in the
// appropriate format.
//
// Supported types include:
//   - nil → serialized as "null"
//   - bool → serialized as "true" or "false"
//   - string → serialized with proper escaping
//   - float64 → serialized as a canonical JSON number
//   - float32, int, int8, int16, int32, int64, uint, uint8, uint16,
//     uint32, uint64 → converted to float64 when within IEEE‑754 safe range
//     (±(2^53 − 1)); otherwise ErrNumberOOR is returned
//   - slices of common types (ints, uints, floats, strings, bools, any)
//   - map[string]any → serialized as a JSON object with keys ordered
//     by UTF‑16 code unit comparison, as required by RFC 8785
//
// Errors:
//   - ErrNumberOOR is returned when an integer cannot be represented
//     exactly in IEEE‑754 double precision.
//   - ErrUnsupportedType is returned when v is of a type not supported
//     by this implementation.
//
// This function is the core entry point for canonical JSON serialization
// in the package. It ensures deterministic output suitable for cryptographic
// operations such as hashing and signing.
// Example:
//
//	 var buf []byte
//	 response := map[string]any{
//		"user_id": "c3f65f70-eb2f-4979-ba73-24bcbde9fdd9",
//		"age":     31,
//	 }
//	 buf, _ = jcs.Append(buf, response)
//	 fmt.Println(string(buf))
//	 Output: {"age":31,"user_id":"c3f65f70-eb2f-4979-ba73-24bcbde9fdd9"}
func Append(dst []byte, v any) ([]byte, error) {
	switch v := v.(type) {
	case nil:
		return append(dst, 'n', 'u', 'l', 'l'), nil

	case bool:
		// Serialize boolean values: "true" or "false"
		if v {
			return append(dst, 't', 'r', 'u', 'e'), nil
		} else {
			return append(dst, 'f', 'a', 'l', 's', 'e'), nil
		}

	case string:
		return appendString(dst, v)

	case float64:
		return appendNumber(dst, v)

	case float32:
		return Append(dst, float64(v))

	case int:
		if isNumberOOR(v) {
			return dst, ErrNumberOOR
		}
		return Append(dst, float64(v))

	case int8:
		return Append(dst, float64(v))

	case int16:
		return Append(dst, float64(v))

	case int32:
		return Append(dst, float64(v))

	case int64:
		if isNumberOOR(v) {
			return dst, ErrNumberOOR
		}

		return Append(dst, float64(v))

	case uint:
		if isNumberOOR(v) {
			return dst, ErrNumberOOR
		}

		return Append(dst, float64(v))

	case uint8:
		return Append(dst, float64(v))

	case uint16:
		return Append(dst, float64(v))

	case uint32:
		return Append(dst, float64(v))

	case uint64:
		if isNumberOOR(v) {
			return dst, ErrNumberOOR
		}

		return Append(dst, float64(v))

	case []int:
		return appendSlice(dst, v)

	case []int8:
		return appendSlice(dst, v)

	case []int16:
		return appendSlice(dst, v)

	case []int32:
		return appendSlice(dst, v)

	case []int64:
		return appendSlice(dst, v)

	case []uint:
		return appendSlice(dst, v)

	case []uint8:
		return appendSlice(dst, v)

	case []uint16:
		return appendSlice(dst, v)

	case []uint32:
		return appendSlice(dst, v)

	case []uint64:
		return appendSlice(dst, v)

	case []any:
		return appendSlice(dst, v)

	case []bool:
		return appendSlice(dst, v)

	case []string:
		return appendSlice(dst, v)

	case []float32:
		return appendSlice(dst, v)

	case []float64:
		return appendSlice(dst, v)

	case time.Time:
		return appendTime(dst, v), nil

	case map[string]any:
		// Go strings are UTF-8
		// RFC 8785 requires UTF-16 code unit comparison, which
		// diffres for non-BMP chars.
		// (e.g., (U+1D11E) → UTF-16 surrogate pair )
		return appendObject(dst, v)
	}

	return dst, ErrUnsupportedType
}
