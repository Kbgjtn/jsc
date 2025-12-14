package jcs

import (
	"unicode/utf16"
	"unicode/utf8"
)

// kv represents a key stored as a UTF-16 encoded string segment.
// It stores the raw string, along with the starting position and length of the key
// in terms of UTF-16 code units (not bytes), which is important for accurate string
// processing and comparisons involving UTF-16 encoded data.
type kv struct {
	// raw holds the original UTF-16 encoded string.
	// It's not necessarily a UTF-16 encoded byte sequence, but a regular Go string
	// interpreted or processed as UTF-16 for the purpose of key management.
	raw string

	// start is the starting position of the key within some larger string or dataset.
	// It helps to know where the key begins when working with a collection of UTF-16 data.
	start int

	// len is the length of the key in terms of UTF-16 code units (i.e., number of 16-bit units).
	// This accounts for surrogate pairs and ensures proper indexing.
	len int
}

// appendUTF16 appends the UTF-16 encoded units of a string `s` to the given buffer `buf`
// and returns the updated buffer, the number of UTF-16 units added, and any error encountered.
//
// this function assume that value of s is a valid UTF-8 string, it cost more when check both rune and the entires,
// and cause slightly overhead at very large sizes, at 100k runes cost nearly doubles compared to small inputs.
// If you expect very large strings, consider chunking or parallelization
func appendUTF16(buf []uint16, s string) ([]uint16, int, error) {
	start := len(buf)

	for _, r := range s {
		if !utf8.ValidRune(r) {
			return buf, 0, ErrInvalidUTF8
		}

		// handle the encoding into UTF-16
		// check if the rune is less than Basic Multilingual Plane (BMP, [U+0000–U+FFFF])
		if r < 0x10_000 {
			// means r can be represented with a single 16-bit value in UTF-16
			buf = append(buf, uint16(r))
		} else {
			// Since UTF-16 uses 16-bit units, it cannot directly represent these
			// higher code points with a single unit. Instead, it uses two 16-bit
			// code units (a high and lower surrogate) to encode one character.
			// basically, we need a way to fit them into two 16-bit units.
			//
			// Surrogate range:
			//  - High surrogates: U+D800–U+DBFF
			//  - Low surrogates: U+DC00–U+DFFF
			// these ranges are reserved exclusively for surrogate use and are not
			// valid standalone characters.

			// substracting into 20 bits padded
			// and split into two 10‑bit halves
			// we get the result of high 10-bits and low 10-bits
			h, l := utf16.EncodeRune(r)
			buf = append(buf, uint16(h), uint16(l))
		}
	}

	return buf, len(buf) - start, nil
}
