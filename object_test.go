package jcs

import (
	"errors"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestAppendObject(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		want    string
		wantErr error
	}{
		{"EmptyObject", map[string]any{}, `{}`, nil},
		{"SingleKeyValue", map[string]any{"foo": "bar"}, `{"foo":"bar"}`, nil},
		{
			"MultipleKeysSorted",
			map[string]any{
				"z": 1,
				"a": true,
				"m": "mid",
			},
			// Keys must be sorted lexicographically: "a","m","z"
			`{"a":true,"m":"mid","z":1}`,
			nil,
		},
		{
			"NestedObject",
			map[string]any{
				"outer": map[string]any{
					"inner": "value",
				},
			},
			`{"outer":{"inner":"value"}}`,
			nil,
		},
		{
			"InvalidUTF8Key",
			map[string]any{
				string([]byte{0xff}): "bad",
			},
			"",
			ErrInvalidUTF8,
		},
		{
			"UnsupportedValueType",
			map[string]any{
				"bad": func() {},
			},
			"",
			ErrUnsupportedType,
		},
		{"ErrUnsupportedType", map[string]any{"err": errors.New("fail")}, "", ErrUnsupportedType},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := appendObject([]byte{}, tc.value)
			Equals(t, tc.wantErr, err)
			Equals(t, tc.want, string(got))
		})
	}
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

func BenchmarkAppendObject(b *testing.B) {
	b.ReportAllocs()

	if testing.Short() {
		b.SkipNow()
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, size := range benchSizes() {
		buf := make([]byte, 0, 1024)
		b.Run(
			"Size"+strconv.Itoa(size),

			func(b *testing.B) {
				dst := buf[:0]
				sample := randomMap(size, rng)

				b.ResetTimer()
				for b.Loop() {
					_, err := appendObject(dst, sample)
					if err != nil {
						b.Fatal(err)
						return
					}
				}
			},
		)
	}
}
