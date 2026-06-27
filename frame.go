package gorle

import (
	"encoding/binary"
	"math"
)

// DecodeFrame decodes a single RLE Lossless frame.
func DecodeFrame(src []byte, nrPixels int, bitsAllocated int, byteOrder ByteOrder) ([]byte, error) {
	bytesPerPixel, err := bytesPerPixel(bitsAllocated, byteOrder)
	if err != nil {
		return nil, err
	}

	if len(src) < 64 {
		return nil, ErrInsufficientData
	}

	var header [64]byte
	copy(header[:], src[:64])
	allOffsets := parseHeaderOffsets(header)
	if allOffsets[0] != 64 {
		return nil, ErrInvalidSegmentOffset
	}

	var offsets []uint32
	for _, val := range allOffsets {
		if val != 0 {
			offsets = append(offsets, val)
		}
	}
	offsets = append(offsets, uint32(len(src)))

	var last uint32
	for _, val := range offsets {
		if val <= last {
			return nil, ErrInvalidSegmentOffset
		}
		last = val
	}

	nrSegments := len(offsets) - 1
	spp := nrSegments / int(bytesPerPixel)
	switch spp {
	case 1:
	case 3:
		if bitsAllocated == 1 {
			return nil, ErrSamplesPerPixelBA1
		}
	default:
		return nil, ErrInvalidSamplesPerPixel
	}

	pps := nrPixels
	expectedLength := nrPixels * nrSegments
	frame := make([]byte, expectedLength)

	for sample := 0; sample < spp; sample++ {
		so := sample * int(bytesPerPixel) * pps
		for byteOffset := 0; byteOffset < int(bytesPerPixel); byteOffset++ {
			idx := segmentIndex(sample, byteOffset, int(bytesPerPixel), spp, byteOrder)
			start := int(offsets[idx])
			end := int(offsets[idx+1])
			segLen, err := decodeSegmentIntoFrame(src[start:end], frame, int(bytesPerPixel), byteOffset+so)
			if err != nil {
				return nil, err
			}
			if segLen != pps {
				return nil, ErrSegmentLength
			}
		}
	}
	return frame, nil
}

// EncodeFrame encodes planar-configuration-0 pixel data (R1,G1,B1,R2,…).
func EncodeFrame(src []byte, rows, cols, spp, bitsAllocated int, byteOrder ByteOrder) ([]byte, error) {
	bytesPerPixel, err := validateEncodeParams(src, rows, cols, spp, bitsAllocated, byteOrder)
	if err != nil {
		return nil, err
	}

	nrSegments := spp * int(bytesPerPixel)
	dst := make([]byte, 64)
	binary.LittleEndian.PutUint32(dst[0:4], uint32(nrSegments))

	startIndices := make([]int, nrSegments)
	for i := range startIndices {
		startIndices[i] = i
	}
	if byteOrder != BigEndian {
		for idx := 0; idx < spp; idx++ {
			s := idx * int(bytesPerPixel)
			e := (idx + 1) * int(bytesPerPixel)
			reverse(startIndices[s:e])
		}
	}

	for idx := 0; idx < nrSegments; idx++ {
		currentOffset := uint32(len(dst))
		binary.LittleEndian.PutUint32(dst[4+idx*4:8+idx*4], currentOffset)

		step := spp * int(bytesPerPixel)
		segment := make([]byte, 0, len(src)/step)
		for i := startIndices[idx]; i < len(src); i += step {
			segment = append(segment, src[i])
		}
		if err := encodeSegmentAppend(&dst, segment, cols); err != nil {
			return nil, err
		}
	}
	return dst, nil
}

func segmentIndex(sample, byteOffset, bytesPerPixel, spp int, byteOrder ByteOrder) int {
	if byteOrder == BigEndian {
		return sample*bytesPerPixel + byteOffset
	}
	return bytesPerPixel - byteOffset + bytesPerPixel*sample - 1
}

func bytesPerPixel(bitsAllocated int, byteOrder ByteOrder) (uint8, error) {
	switch bitsAllocated {
	case 0:
		return 0, ErrInvalidBitsAllocated
	case 1:
		return 1, nil
	default:
		if bitsAllocated%8 != 0 {
			return 0, ErrInvalidBitsAllocated
		}
		bpp := bitsAllocated / 8
		if bpp == 1 {
			return 1, nil
		}
		if bpp != 2 && bpp != 4 && bpp != 8 {
			return 0, ErrInvalidBitsAllocated
		}
		if byteOrder != LittleEndian && byteOrder != BigEndian {
			return 0, ErrInvalidByteOrder
		}
		return uint8(bpp), nil
	}
}

func validateEncodeParams(src []byte, rows, cols, spp, bitsAllocated int, byteOrder ByteOrder) (uint8, error) {
	var bytesPerPixel uint8
	switch spp {
	case 1:
		switch bitsAllocated {
		case 1:
			bytesPerPixel = 1
		case 8, 16, 32, 64:
			bytesPerPixel = uint8(bitsAllocated / 8)
		default:
			return 0, ErrInvalidBitsAllocated
		}
	case 3:
		switch bitsAllocated {
		case 1:
			return 0, ErrSamplesPerPixelBA1
		case 8, 16, 32, 64:
			bytesPerPixel = uint8(bitsAllocated / 8)
		default:
			return 0, ErrInvalidBitsAllocated
		}
	default:
		return 0, ErrInvalidSamplesPerPixel
	}

	if bytesPerPixel > 1 {
		if byteOrder != LittleEndian && byteOrder != BigEndian {
			return 0, ErrInvalidByteOrder
		}
	}

	totalPixels := rows * cols * spp
	totalLength := totalPixels * int(bytesPerPixel)
	if len(src) != totalLength {
		return 0, ErrInvalidParameters
	}
	if spp*int(bytesPerPixel) > 15 {
		return 0, ErrTooManySegments
	}
	return bytesPerPixel, nil
}

func encodeSegmentAppend(dst *[]byte, segment []byte, cols int) error {
	if cols <= 0 || len(segment)%cols != 0 {
		return ErrInvalidColumns
	}
	for offset := 0; offset < len(segment); offset += cols {
		if err := encodeRow(nil, segment[offset:offset+cols], dst); err != nil {
			return err
		}
	}
	if len(*dst)%2 != 0 {
		*dst = append(*dst, 0)
	}
	return nil
}

func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// UnpackedFrameLength returns the decoded frame length in bytes for validation.
func UnpackedFrameLength(rows, cols, spp, bitsAllocated int) int {
	bpp := bitsAllocated
	if bpp > 1 {
		bpp = bitsAllocated / 8
	}
	return rows * cols * spp * bpp
}

// Packed1BitLength returns packed 1-bit frame size in bytes.
func Packed1BitLength(rows, cols int) int {
	return int(math.Ceil(float64(rows*cols) / 8))
}
