package jcs

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestAppendUTF16(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []uint16
		wantErr error
	}{
		{"ascii characters", "A", []uint16{0x41}, nil},
		{"Latin character", "é", []uint16{0xE9}, nil},
		{"Euro sign", "€", []uint16{0x20AC}, nil},
		{"U+1F600", "😀", []uint16{0xD83D, 0xDE00}, nil},
		{"U+FFFD", string([]byte{0xEF, 0xBF, 0xBD}), []uint16{0xfffd}, nil}, // replacement character
		{"mixed string", "😀€éA", []uint16{0xD83D, 0xDE00, 0x20AC, 0xE9, 0x41}, nil},
		{"invalid single byte rune", string([]byte{0xFF}), make([]uint16, 0), ErrInvalidUTF8},
		{"truncated 3-byte sequence", string([]byte{0xE2, 0x82}), make([]uint16, 0), ErrInvalidUTF8},
		{"Truncated4Byte", string([]byte{0xF0, 0x9F, 0x92}), make([]uint16, 0), ErrInvalidUTF8},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := make([]uint16, 0, len(tc.want))
			got, n, err := appendUTF16(buf, tc.in)
			Equals(t, tc.wantErr, err)
			Equals(t, tc.want, got)
			Equals(t, len(tc.want), n)
		})
	}
}

func BenchmarkAppendUTF16(b *testing.B) {
	b.ReportAllocs()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, size := range benchSizes() {
		buf := make([]uint16, 0, size*2)
		s := randomString(size, rng)

		b.Run(
			"Size-"+strconv.Itoa(size),
			func(b *testing.B) {
				dst := buf[:0]

				b.ResetTimer()
				for b.Loop() {
					_, _, err := appendUTF16(dst, s)
					if err != nil {
						b.Fatal(err)
						return
					}
				}
			},
		)
	}
}
