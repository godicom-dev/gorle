package gorle

// ByteOrder is the sample byte order for multi-byte pixels.
type ByteOrder string

const (
	LittleEndian ByteOrder = "<"
	BigEndian    ByteOrder = ">"
)

// PixelDataVersion selects decode_pixel_data behaviour.
type PixelDataVersion int

const (
	PixelDataV1 PixelDataVersion = 1
	PixelDataV2 PixelDataVersion = 2
)

// FrameOptions configures single-frame RLE encode/decode.
type FrameOptions struct {
	Rows            int
	Columns         int
	SamplesPerPixel int
	BitsAllocated   int
	ByteOrder       ByteOrder
}

// PixelDataOptions configures DecodePixelData / EncodePixelData.
type PixelDataOptions struct {
	Version   PixelDataVersion
	FrameOptions
	PackBits bool // v2 only: return 1-bit data packed when BitsAllocated is 1
}
