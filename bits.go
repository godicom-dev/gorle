package gorle

// PackBits packs binary data (0/1 bytes) into one bit per pixel.
func PackBits(src []byte, byteOrder ByteOrder) ([]byte, error) {
	if byteOrder != LittleEndian && byteOrder != BigEndian {
		return nil, ErrInvalidByteOrder
	}
	for _, b := range src {
		if b > 1 {
			return nil, ErrInvalidPackBitsInput
		}
	}

	dst := make([]byte, 0, (len(src)+7)/8)
	for offset := 0; offset+8 <= len(src); offset += 8 {
		dst = append(dst, packByte(src[offset:offset+8], byteOrder))
	}

	remainder := len(src) % 8
	if remainder == 0 {
		return dst, nil
	}

	tail := src[len(src)-remainder:]
	var last byte
	if byteOrder == LittleEndian {
		for idx, bit := range tail {
			last |= bit << idx
		}
	} else {
		for idx, bit := range tail {
			last |= bit << (7 - idx)
		}
	}
	return append(dst, last), nil
}

// UnpackBits unpacks bit-packed data. count is the number of bits to return;
// zero means unpack the entire input.
func UnpackBits(src []byte, count int, byteOrder ByteOrder) ([]byte, error) {
	if byteOrder != LittleEndian && byteOrder != BigEndian {
		return nil, ErrInvalidByteOrder
	}

	maxBits := len(src) * 8
	if count <= 0 || count > maxBits {
		count = maxBits
	}

	nrWholeBytes := count / 8
	nrRemainderBits := count % 8
	dst := make([]byte, 0, count)

	if byteOrder == LittleEndian {
		for offset := 0; offset < nrWholeBytes; offset++ {
			for idx := 0; idx < 8; idx++ {
				dst = append(dst, (src[offset]>>idx)&1)
			}
		}
		for idx := 0; idx < nrRemainderBits; idx++ {
			dst = append(dst, (src[nrWholeBytes]>>idx)&1)
		}
	} else {
		for offset := 0; offset < nrWholeBytes; offset++ {
			for idx := 0; idx < 8; idx++ {
				dst = append(dst, (src[offset]>>(7-idx))&1)
			}
		}
		for idx := 0; idx < nrRemainderBits; idx++ {
			dst = append(dst, (src[nrWholeBytes]>>(7-idx))&1)
		}
	}
	return dst, nil
}

func packByte(chunk []byte, byteOrder ByteOrder) byte {
	var packed byte
	if byteOrder == LittleEndian {
		for idx, bit := range chunk {
			packed |= bit << idx
		}
	} else {
		for idx, bit := range chunk {
			packed |= bit << (7 - idx)
		}
	}
	return packed
}
