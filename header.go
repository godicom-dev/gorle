package gorle

import (
	"encoding/binary"
)

// ParseHeader returns the 15 segment offsets from a 64-byte RLE header.
// Bytes 0–3 hold the segment count; offsets are little-endian uint32 values
// starting at byte 4 (DICOM PS3.5 G.3).
func ParseHeader(src []byte) ([]uint32, error) {
	if len(src) != 64 {
		return nil, ErrInvalidHeaderLength
	}
	offsets := make([]uint32, 15)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(src[4+i*4 : 8+i*4])
	}
	return offsets, nil
}

func parseHeaderOffsets(src [64]byte) [15]uint32 {
	var offsets [15]uint32
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(src[4+i*4 : 8+i*4])
	}
	return offsets
}
