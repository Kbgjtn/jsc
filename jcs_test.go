package jcs

import (
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

var (
	bigSizesBench   = []int{100, 1000, 10000, 100_000, 1_000_000, 5_000_000, 10_000_000}
	smallSizesBench = []int{100, 1_000, 10_000, 15_000, 25_000, 50_000, 100_000}
)

func benchSizes() []int {
	if testing.Short() {
		return smallSizesBench
	}
	return bigSizesBench
}

func randomString(n int, rng *rand.Rand) string {
	runes := make([]rune, n)

	for i := 0; i < n; i++ {
		switch rng.Intn(3) {
		// ASCII
		case 0:
			runes[i] = rune(rng.Intn(0x7F))

		// BMP
		case 1:
			runes[i] = rune(0x80 + rng.Intn(0x7FF-0x80))

		// Supplementary plane
		case 2:
			runes[i] = rune(0x10_000 + rng.Intn(0x10FFFF-0x10_000))
		}
	}

	return string(runes)
}

// randomMap generates a map[string]any with n entries.
// Keys are random strings, values are random types (string, int, bool, float64).
func randomMap(n int, rng *rand.Rand) map[string]any {
	m := make(map[string]any, n)

	for i := 0; i < n; i++ {
		// Generate a random key
		key := "key" + strconv.Itoa(rng.Intn(1_000_000))

		// Randomize value type
		switch rng.Intn(4) {
		case 0:
			m[key] = rng.Intn(1000) // int
		case 1:
			m[key] = rng.Float64() * 1000 // float64
		case 2:
			m[key] = "val" + strconv.Itoa(rng.Intn(1_000)) // string
		case 3:
			m[key] = rng.Intn(2) == 0 // bool
		}
	}

	return m
}

func Equals(tb testing.TB, expected, actual any) {
	tb.Helper()

	if !reflect.DeepEqual(expected, actual) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, expected, actual)
		tb.FailNow()
	}
}

// Example shows how to use Append in documentation.
func Example() {
	var buf []byte

	response := map[string]any{
		"user_id": "c3f65f70-eb2f-4979-ba73-24bcbde9fdd9",
		"age":     31,
	}

	// Append a string
	buf, _ = Append(buf, response)
	fmt.Println(string(buf))
	// Output: {"age":31,"user_id":"c3f65f70-eb2f-4979-ba73-24bcbde9fdd9"}
}

func TestAppend(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    string
		wantErr error
	}{
		{name: "Empty object", value: map[string]any{}, want: `{}`},
		{name: "Key ordering", value: map[string]any{"b": 1, "a": 2}, want: `{"a":2,"b":1}`},
		{name: "String escaping newline", value: map[string]any{"s": "line\nbreak"}, want: `{"s":"line\nbreak"}`},
		{name: "Quote and backslash escaping", value: map[string]any{"s": `a"b\c`}, want: "{\"s\":\"a\\\"b\\\\c\"}"},
		{name: "Control character U+0001", value: map[string]any{"s": string(rune(1))}, want: "{\"s\":\"\\u0001\"}"},
		{name: "Non-ASCII character", value: map[string]any{"s": "é"}, want: `{"s":"é"}`},
		{name: "Number normalization integer", value: map[string]any{"n": 1.0}, want: `{"n":1}`},
		{name: "Number normalization exponent", value: map[string]any{"n": 1e+00}, want: `{"n":1}`},
		{name: "Negative zero", value: map[string]any{"n": -0.0}, want: `{"n":0}`},
		{name: "Large exponent", value: map[string]any{"n": 1e-6}, want: `{"n":0.000001}`},
		{name: "String subtype time", value: map[string]any{"time": "2019-01-28T07:45:10Z"}, want: `{"time":"2019-01-28T07:45:10Z"}`},
		{name: "Key collision lookalike", value: map[string]any{"a": 1, "u0061": 2}, want: `{"a":1,"u0061":2}`},
		{name: "Array canonicalization", value: []any{1.0, "a", true}, want: `[1,"a",true]`},
		{name: "Surrogate Pairs", value: map[string]any{"s": "😀"}, want: `{"s":"😀"}`},
		{name: "Key Ordering with Mixed Cases", value: map[string]any{"A": 1, "a": 2}, want: `{"A":1,"a":2}`},     // "A" (U+0041) < "a" (U+0061).
		{name: "Key Ordering with Surrogate Pairs", value: map[string]any{"😀": 1, "a": 2}, want: `{"a":2,"😀":1}`}, // U+0061 < U+1F600.
		{name: "Nested Objects", value: map[string]any{"outer": map[string]any{"b": 1, "a": 2}}, want: `{"outer":{"a":2,"b":1}}`},
		{name: "Nested Arrays", value: []any{[]any{2, 1}, 3}, want: `[[2,1],3]`},
		{name: "Boolean values", value: map[string]any{"t": true, "f": false}, want: `{"f":false,"t":true}`},
		{name: "Null values", value: map[string]any{"n": nil}, want: `{"n":null}`},
		{name: "Large integers Int", value: map[string]any{"n": int(1<<53 - 1)}, want: `{"n":9007199254740991}`},
		{name: "Large integers Int64", value: map[string]any{"n": int64(1<<53 - 1)}, want: `{"n":9007199254740991}`},
		{name: "Large integers Int8", value: map[string]any{"n": int8(math.MaxInt8)}, want: `{"n":127}`},
		{name: "Large integers Int16", value: map[string]any{"n": int16(math.MaxInt16)}, want: `{"n":32767}`},
		{name: "Large integers Int32", value: map[string]any{"n": int32(math.MaxInt32)}, want: `{"n":2147483647}`},
		{name: "Large negative integers Int", value: map[string]any{"n": -int(1<<53 - 1)}, want: `{"n":-9007199254740991}`},
		{name: "Large negative integers Int64", value: map[string]any{"n": -int64(1<<53 - 1)}, want: `{"n":-9007199254740991}`},
		{name: "Large negative integers Int8", value: map[string]any{"n": -int8(math.MaxInt8)}, want: `{"n":-127}`},
		{name: "Large negative integers Int16", value: map[string]any{"n": -int16(math.MaxInt16)}, want: `{"n":-32767}`},
		{name: "Large negative integers Int32", value: map[string]any{"n": -int32(math.MaxInt32)}, want: `{"n":-2147483647}`},
		{name: "Large integers Uint", value: map[string]any{"n": uint(1<<53 - 1)}, want: `{"n":9007199254740991}`},
		{name: "Large integers Uint64", value: map[string]any{"n": uint64(1<<53 - 1)}, want: `{"n":9007199254740991}`},
		{name: "Large integers Uint8", value: map[string]any{"n": uint8(math.MaxUint8)}, want: `{"n":255}`},
		{name: "Large integers Uint16", value: map[string]any{"n": uint16(math.MaxUint16)}, want: `{"n":65535}`},
		{name: "Large integers Uint32", value: map[string]any{"n": uint32(math.MaxUint32)}, want: `{"n":4294967295}`},
		{name: "ErrInvalidUTF8", value: map[string]any{"s": string([]byte{0xff})}, want: "", wantErr: ErrInvalidUTF8},
		{name: "ErrNumberOutOfRange_int", value: map[string]any{"n": math.MaxInt64}, wantErr: ErrNumberOOR},
		{name: "ErrNumberOutOfRange_int64", value: map[string]any{"n": int64(math.MaxInt64)}, wantErr: ErrNumberOOR},
		{name: "ErrNumberOutOfRange_uint", value: map[string]any{"n": uint(math.MaxUint)}, wantErr: ErrNumberOOR},
		{name: "ErrNumberOutOfRange_uint64", value: map[string]any{"n": uint64(math.MaxUint64)}, wantErr: ErrNumberOOR},
		{name: "ErrNan", value: map[string]any{"n": math.NaN()}, wantErr: ErrNaN},
		{name: "ErrInf", value: map[string]any{"n": math.Inf(-0)}, wantErr: ErrInf},
		{name: "ErrInf-", value: map[string]any{"n": math.Inf(-0)}, wantErr: ErrInf},
		{name: "-0", value: map[string]any{"n": -0}, want: `{"n":0}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := Append(nil, tc.value)
			Equals(t, tc.wantErr, err)
			Equals(t, tc.want, string(out))
		})
	}
}
