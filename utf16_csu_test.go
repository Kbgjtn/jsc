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
		{name: "ascii characters", in: "A", want: []uint16{0x41}},
		{name: "Latin character", in: "é", want: []uint16{0xE9}},
		{name: "Euro sign", in: "€", want: []uint16{0x20AC}},
		{name: "U+1F600", in: "😀", want: []uint16{0xD83D, 0xDE00}},
		{name: "U+FFFD", in: string([]byte{0xEF, 0xBF, 0xBD}), want: []uint16{0xfffd}}, // replacement character
		{name: "mixed string", in: "😀€éA", want: []uint16{0xD83D, 0xDE00, 0x20AC, 0xE9, 0x41}},
		{name: "invalid single byte rune", in: string([]byte{0xFF}), want: make([]uint16, 0), wantErr: ErrInvalidUTF8},
		{name: "truncated 3-byte sequence", in: string([]byte{0xE2, 0x82}), want: make([]uint16, 0), wantErr: ErrInvalidUTF8},
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

func BenchmarkAppendUTF16(b *testing.B) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var sizes []int = globalSizes
	if testing.Short() {
		sizes = []int{100, 1_000, 10_000, 15_000, 25_000, 50_000, 100_000}
	}

	for _, size := range sizes {
		buf := make([]uint16, 0, size*2)
		s := randomString(size, rng)

		b.Run(
			"Size-"+strconv.Itoa(size),
			func(b *testing.B) {
				dst := buf[:0]

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
