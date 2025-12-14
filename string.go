package jcs

import (
	"slices"
	"unicode/utf8"
)

var hex = []byte("0123456789abcdef")

// appendString appends the canonical JSON representation of a Go string to dst.
//
// This function implements the string escaping and UTF-8 validation rules
// required by RFC 8785 (JSON Canonicalization Scheme):
//
//   - All output is UTF-8 encoded and enclosed in double quotes.
//   - Safe ASCII characters (U+0020–U+007E, excluding '"' and '\') are copied
//     directly for efficiency.
//   - Control characters (< U+0020) and the special characters '"' and '\' are
//     escaped using JSON escape sequences (e.g., \u00XX, \n, \t).
//   - Non-ASCII characters are validated and emitted as-is.
//
// UTF-8 validation:
//   - utf8.DecodeRuneInString is used to decode each rune.
//   - If DecodeRune returns utf8.RuneError, both invalid single-byte sequences
//     and truncated multi-byte sequences are rejected with ErrInvalidUTF8.
//   - A literal U+FFFD replacement character (encoded as 0xEF 0xBF 0xBD) is
//     allowed, since it is a valid Unicode scalar value.
//   - Surrogate code points (U+D800–U+DFFF) are explicitly rejected, as they
//     are not valid Unicode scalar values and disallowed by RFC 8785.
//
// Error handling:
//   - Returns ErrInvalidUTF8 if the input string contains malformed UTF-8 or
//     surrogate code points.
//   - Otherwise, returns the updated dst slice containing the escaped string.
//
// The resulting output is guaranteed to be a valid, canonical JSON string
// according to RFC 8785.
func appendString(dst []byte, s string) ([]byte, error) {
	dstLen := len(dst)

	// Worst case: every byte escaped + quotes
	dst = slices.Grow(dst, len(s)+2)

	dst = append(dst, '"')

	for i := 0; i < len(s); {
		c := s[i]

		// Fast path: copy contiguous safe ASCII
		if c < utf8.RuneSelf && c >= 0x20 && c != '"' && c != '\\' {
			start := i
			i++

			for i < len(s) {
				c = s[i]
				if c < utf8.RuneSelf && c >= 0x20 && c != '"' && c != '\\' {
					i++
					continue
				}
				break
			}
			dst = append(dst, s[start:i]...)
			continue
		}

		// ASCII slow path (escaping)
		if c < utf8.RuneSelf {
			switch c {
			case '"', '\\':
				dst = append(dst, '\\', c)
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\t':
				dst = append(dst, '\\', 't')
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\f':
				dst = append(dst, '\\', 'f')
			case '\r':
				dst = append(dst, '\\', 'r')
			default:
				// control character → \u00XX
				dst = append(dst, '\\', 'u', '0', '0',
					hex[c>>4],
					hex[c&0xF],
				)
			}
			i++
			continue
		}

		// Non-ASCII: validate UTF-8 and emit bytes as-is
		r, size := utf8.DecodeRuneInString(s[i:])

		// catches and reject both single-byte and truncated multi-byte invalids sequences.
		if r == utf8.RuneError {
			if size == 3 && s[i] == 0xEF && s[i+1] == 0xBF && s[i+2] == 0xBD {
				// valid U+FFFD, allow
			} else {
				return dst[:dstLen], ErrInvalidUTF8
			}
		}

		// reject surrogate code points (U+D800–U+DFFF)
		if r >= 0xD800 && r <= 0xDFFF {
			return dst[:dstLen], ErrInvalidUTF8
		}

		dst = append(dst, s[i:i+size]...)
		i += size
	}

	dst = append(dst, '"')
	return dst, nil
}
