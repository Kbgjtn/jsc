package jcs

import "sort"

// appendObject serializes a map[string]any (JSON object) into the destination byte slice `dst`.
// The function sorts the keys lexicographically, processes UTF-16 encoding for key/value pairs,
// and ensures the JSON object is serialized in canonical form as per RFC 8785.
func appendObject(dst []byte, obj map[string]any) ([]byte, error) {
	dstLen := len(dst)
	dst = append(dst, '{')
	if len(obj) == 0 {
		return append(dst, '}'), nil
	}

	// shared UTF-16 buffer for all keys
	// heuristic: ASCII keys dominate - ~1 code unit per byte
	utf16buf := make([]uint16, 0, len(obj)*8)

	keys := make([]kv, 0, len(obj))
	for k := range obj {
		start := len(utf16buf)
		var n int
		var err error

		utf16buf, n, err = appendUTF16(utf16buf, k)
		if err != nil {
			return dst[:dstLen], err
		}

		keys = append(keys, kv{
			raw:   k,
			len:   n,
			start: start,
		})

	}

	sort.Slice(keys, func(i, j int) bool {
		a := utf16buf[keys[i].start : keys[i].start+keys[i].len]
		b := utf16buf[keys[j].start : keys[j].start+keys[j].len]

		n := min(len(a), len(b))

		for k := 0; k < n; k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return len(a) < len(b)
	})

	for i, k := range keys {
		if i > 0 {
			dst = append(dst, ',')
		}

		var err error

		// key
		dst, err = appendString(dst, k.raw)
		if err != nil {
			return dst[:dstLen], err
		}

		dst = append(dst, ':')
		dst, err = Append(dst, obj[k.raw])
		if err != nil {
			return dst[:dstLen], err
		}
	}

	dst = append(dst, '}')
	return dst, nil
}
