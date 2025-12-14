package jcs

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestAppendString(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr error
	}{
		{"Empty", "", `""`, nil},
		{"SimpleASCII", "simple ascii", `"simple ascii"`, nil},
		{"QuoteEscape", `quote"slash\test`, `"quote\"slash\\test"`, nil},
		{"ControlChars", "control:\b\t\n\f\r", `"control:\b\t\n\f\r"`, nil},
		{"LowUnicode", "low\u0001\u0002\u001f", `"low\u0001\u0002\u001f"`, nil},
		{"NonASCII", "こんにちは世界", `"こんにちは世界"`, nil},
		{"Emoji", "emoji 😀😅🚀", `"emoji 😀😅🚀"`, nil},
		{"Mixed", "mixed ascii 日本語 😀 \n \" \\", `"mixed ascii 日本語 😀 \n \" \\"`, nil},
		{"U+FFFD valid", string([]byte{0xEF, 0xBF, 0xBD}), string("\"\uFFFD\""), nil}, // replacement character
		{"Invalid single-byte sequences", string([]byte{0xFF}), "", ErrInvalidUTF8},   // invalid UTF-8 byte
		{"truncated 3-byte sequence", string([]byte{0xE2, 0x82}), "", ErrInvalidUTF8},
		{"SurrogateHigh", string([]byte{0xED, 0xA0, 0x80}), "", ErrInvalidUTF8},
		{"SurrogateLow", string([]byte{0xED, 0xBF, 0xBF}), "", ErrInvalidUTF8},
		{"OverlongEncoding", string([]byte{0xC0, 0x80}), "", ErrInvalidUTF8},
		{"Truncated4Byte", string([]byte{0xF0, 0x9F, 0x92}), "", ErrInvalidUTF8},
		{"InvalidContinuation", string([]byte{0xE2, 0x28, 0xA1}), "", ErrInvalidUTF8},
		{"OutOfRange", string([]byte{0xF4, 0x90, 0x80, 0x80}), "", ErrInvalidUTF8},
		{"TrailingBackslash", "ends with \\", `"ends with \\"`, nil},
		{"TrailingQuote", "ends with \"", `"ends with \""`, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := appendString([]byte{}, tc.value)
			Equals(t, tc.wantErr, err)
			Equals(t, tc.want, string(out))
		})
	}
}

func BenchmarkAppendString(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, size := range benchSizes() {
		buf := make([]byte, 0, size*3)
		s := randomString(size, rng)

		b.Run(
			"Size-"+strconv.Itoa(size),
			func(b *testing.B) {
				dst := buf[:0]

				for b.Loop() {
					_, err := appendString(dst, s)
					if err != nil {
						b.Fatal(err)
						return
					}
				}
			},
		)
	}
}
