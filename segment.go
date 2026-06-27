package gorle

// DecodeSegment decodes a single RLE segment (PackBits).
func DecodeSegment(src []byte) ([]byte, error) {
	dst := make([]byte, 0, len(src))
	if err := decodeSegmentInto(&dst, src); err != nil {
		return nil, err
	}
	return dst, nil
}

// EncodeSegment encodes a segment. cols is the number of values per row.
func EncodeSegment(src []byte, cols int) ([]byte, error) {
	if cols <= 0 || len(src)%cols != 0 {
		return nil, ErrInvalidColumns
	}
	dst := make([]byte, 0, len(src))
	rowLen := cols
	for offset := 0; offset < len(src); offset += rowLen {
		if err := encodeRow(nil, src[offset:offset+rowLen], &dst); err != nil {
			return nil, err
		}
	}
	if len(dst)%2 != 0 {
		dst = append(dst, 0)
	}
	return dst, nil
}

// EncodeRow RLE-encodes a single row of sample values.
func EncodeRow(src []byte) ([]byte, error) {
	dst := make([]byte, 0, len(src)*2)
	if err := encodeRow(nil, src, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}

func decodeSegmentInto(dst *[]byte, src []byte) error {
	if len(src) == 0 {
		return nil
	}
	pos := 0
	maxOffset := len(src) - 1
	for {
		headerByte := int(src[pos])
		pos++
		switch {
		case headerByte > 128:
			opLen := 257 - headerByte
			if pos > maxOffset {
				return ErrSegmentDecodeEOF
			}
			val := src[pos]
			pos++
			for range opLen {
				*dst = append(*dst, val)
			}
		case headerByte < 128:
			opLen := headerByte + 1
			if pos+headerByte > maxOffset {
				return ErrSegmentDecodeEOF
			}
			*dst = append(*dst, src[pos:pos+opLen]...)
			pos += opLen
		}
		if pos >= maxOffset {
			return nil
		}
	}
}

func decodeSegmentIntoFrame(src []byte, frame []byte, bpp, initialOffset int) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}
	idx := initialOffset
	pos := 0
	maxOffset := len(src) - 1
	maxFrame := len(frame)

	for {
		headerByte := int(src[pos])
		pos++
		switch {
		case headerByte > 128:
			opLen := 257 - headerByte
			if pos > maxOffset {
				return 0, ErrSegmentDecodeEOF
			}
			val := src[pos]
			pos++
			if idx+opLen > maxFrame {
				for range maxFrame - idx {
					frame[idx] = val
					idx += bpp
				}
				return (idx - initialOffset) / bpp, nil
			}
			for range opLen {
				frame[idx] = val
				idx += bpp
			}
		case headerByte < 128:
			opLen := headerByte + 1
			if pos+headerByte > maxOffset {
				return 0, ErrSegmentDecodeEOF
			}
			if idx+opLen > maxFrame {
				for ii := pos; ii < pos+(maxFrame-idx)/bpp; ii++ {
					frame[idx] = src[ii]
					idx += bpp
				}
				return (idx - initialOffset) / bpp, nil
			}
			for ii := pos; ii < pos+opLen; ii++ {
				frame[idx] = src[ii]
				idx += bpp
			}
			pos += opLen
		}
		if pos >= maxOffset {
			return (idx - initialOffset) / bpp, nil
		}
	}
}

func encodeRow(_ []byte, src []byte, dst *[]byte) error {
	switch len(src) {
	case 0:
		return nil
	case 1:
		*dst = append(*dst, 0, src[0])
		return nil
	}

	literal := make([]byte, 0, 128)
	groups := chunkEqualRuns(src)
	for _, group := range groups {
		if len(group) == 1 {
			literal = append(literal, group[0])
			continue
		}
		if len(literal) > 0 {
			appendLiteralRuns(dst, literal)
			literal = literal[:0]
		}
		for offset := 0; offset < len(group); offset += 128 {
			chunk := group[offset:min(offset+128, len(group))]
			if len(chunk) > 1 {
				*dst = append(*dst, byte(257-len(chunk)), chunk[0])
			} else {
				*dst = append(*dst, 0, chunk[0])
			}
		}
	}
	if len(literal) > 0 {
		appendLiteralRuns(dst, literal)
	}
	return nil
}

func appendLiteralRuns(dst *[]byte, literal []byte) {
	for offset := 0; offset < len(literal); offset += 128 {
		chunk := literal[offset:min(offset+128, len(literal))]
		*dst = append(*dst, byte(len(chunk)-1))
		*dst = append(*dst, chunk...)
	}
}

func chunkEqualRuns(src []byte) [][]byte {
	if len(src) == 0 {
		return nil
	}
	var groups [][]byte
	start := 0
	for i := 1; i <= len(src); i++ {
		if i == len(src) || src[i] != src[start] {
			groups = append(groups, src[start:i])
			start = i
		}
	}
	return groups
}
