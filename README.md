## JSON Canonicalization Scheme (JCS) in Go

### Overview

The jcs package implements the JSON Canonicalization Scheme (JCS) as defined in RFC 8785. JCS is a method for converting JSON data into a canonical form, ensuring that the data is consistently represented in a way that allows for reliable comparisons and signatures. This package provides a Go implementation that allows encoding Go values into canonical JSON format, which can be used for digital signatures, data integrity checks, and other cryptographic applications.

### Supported Types and Behavior

1. **`nil`**
   The value `nil` is serialized as `null`.

2. **`bool`**
   The value is serialized as either `"true"` or `"false"`.

3. **`string`**
   The string is serialized using UTF-8 encoding.

4. **Numeric types**
   The following numeric types are supported and are serialized as JSON numbers (with conversion to `float64` where necessary):
   - `float64`
   - `float32` (converted to `float64`)
   - `int` (converted to `float64`)
   - `int8`, `int16`, `int32`, `int64` (converted to `float64`)
   - `uint`, `uint8`, `uint16`, `uint32`, `uint64` (converted to `float64`)

   For `int` and `uint` types that exceed the supported range for JSON numbers, the function will return the `ErrNumberOOR` error.
   This means only integers in the range `[‑(2^53‑1), v ,+(2^53‑1)]` are valid.

5. **Arrays and Slices**
   Slices of basic types (e.g., `[]int`, `[]float64`, `[]string`, etc.) are recursively serialized as arrays in JSON format. Supported slice types include:
   - `[]int`, `[]int8`, `[]int16`, `[]int32`, `[]int64`
   - `[]uint`, `[]uint8`, `[]uint16`, `[]uint32`, `[]uint64`
   - `[]bool`
   - `[]string`
   - `[]float32`, `[]float64`
   - `[]any`

   Each element of the slice is serialized individually, and the resulting canonicalized representation is appended to `dst`.

6. **`time.Time`**
   A `time.Time` value is serialized in the RFC 3339 format (i.e., `2006-01-02T15:04:05Z07:00`).

7. **`map[string]any` (Objects)**
   A `map` is serialized as a JSON object. The keys are encoded as UTF-8 strings, and the values are serialized according to their types. Note that RFC 8785 requires the use of **UTF-16 code unit comparison**, which affects how non-BMP characters (e.g., Unicode surrogate pairs) are handled.

8. **Unsupported Types**
   If the value `v` is of an unsupported type, the function returns the error `ErrUnsupportedType`.

### Error Handling

During the process of encoding Go values into canonical JSON format, various errors can arise based on the type or characteristics of the data being encoded. This section outlines the possible errors that may be returned by the package, helping you understand how to handle them when using the package.

The following errors are defined in the `jcs` package. Each error corresponds to a specific issue that might occur during the canonicalization process:

---

#### 1. `ErrUnsupportedType`

**Description**:
This error occurs when the encoder encounters a value of an unsupported type. The `jcs` encoder supports only a subset of Go types, including basic types like integers, strings, booleans, and more complex types like slices and maps. Custom structs, channels, and function types (among others) are **not supported** by JCS and will trigger this error.

**Possible Causes**:

- Attempting to encode unsupported types such as:
  - Function types
  - Channels
  - Structured types that are not explicitly handled by the encoder
  - Other types than `map[string]interface{}`

- Unsupported composite types that cannot be serialized into a canonical JSON format.

#### 2. `ErrNaN`

**Description**:
The `ErrNaN` error is returned when the encoder encounters a `NaN` (Not a Number) value. According to RFC 8785, `NaN` values are **not allowed** in the canonical JSON format. If a `NaN` value is passed to the encoder, it triggers this error.

**Possible Causes**:

- Attempting to encode `NaN` values, which are invalid in the context of JCS and canonical JSON.

#### 3. `ErrInf`

**Description**:
This error is returned when the encoder encounters an `Inf` (infinity) value, either positive infinity (`+Inf`) or negative infinity (`-Inf`). As per RFC 8785, both `Inf` and `-Inf` are **disallowed** in the canonical JSON format. The encoder will reject such values by returning this error.

**Possible Causes**:

- Attempting to encode positive or negative infinity values.

#### 4. `ErrInvalidUTF8`

**Description**:
The `ErrInvalidUTF8` error is returned when the encoder encounters a string that contains invalid UTF-8 byte sequences. JCS requires that all strings be valid UTF-8. If a string with invalid UTF-8 characters is passed, this error is triggered.

**Possible Causes**:

- Strings containing invalid or corrupted UTF-8 byte sequences.
- Encodings that are not valid UTF-8 (e.g., UTF-16 or other non-UTF-8 encodings).

#### 5. `ErrNumberOOR`

**Description**:
The `ErrNumberOOR` (Out of Range) error is returned when a number exceeds the valid range for precise representation in the IEEE-754 double-precision format. According to JCS and RFC 8785, numbers larger than ±2^53 cannot be represented precisely and will result in this error. This is due to the fact that JCS requires **exact round-trip encoding** of numerical values.

**Possible Causes**:

- Numbers that exceed the precision limits of IEEE-754 double-precision floating-point numbers (approximately ±9.007 x 10^15).
- This applies to both integer and floating-point numbers.
