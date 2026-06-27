package gorle

import "errors"

var (
	ErrInvalidHeaderLength   = errors.New("the RLE header must be 64 bytes long")
	ErrInvalidSegmentOffset  = errors.New("invalid segment offset found in the RLE header")
	ErrInsufficientData      = errors.New("frame is not long enough to contain RLE encoded data")
	ErrInvalidBitsAllocated  = errors.New("the (0028,0100) 'Bits Allocated' value must be 1, 8, 16, 32 or 64")
	ErrInvalidSamplesPerPixel = errors.New("the (0028,0002) 'Samples per Pixel' must be 1 or 3")
	ErrSamplesPerPixelBA1    = errors.New("the (0028,0002) 'Samples per Pixel' must be 1 if (0028,0100) 'Bits Allocated' is 1")
	ErrSegmentLength         = errors.New("the decoded segment length does not match the expected length")
	ErrInvalidByteOrder      = errors.New("'byteorder' must be '>' or '<'")
	ErrSegmentDecodeEOF      = errors.New("the end of the data was reached before the segment was completely decoded")
	ErrTooManySegments       = errors.New("unable to encode as the DICOM Standard only allows a maximum of 15 segments in RLE encoded data")
	ErrInvalidParameters     = errors.New("the length of the data to be encoded is not consistent with the values of the dataset's 'Rows', 'Columns', 'Samples per Pixel' and 'Bits Allocated' elements")
	ErrInvalidColumns        = errors.New("the (0028,0011) 'Columns' value is invalid")
	ErrInvalidPackBitsInput  = errors.New("only binary input (containing zeros or ones) can be packed")
	ErrMissingPixelDataArgs   = errors.New("missing expected keyword arguments: bits_allocated, columns, rows")
	ErrUnsupportedPixelVersion = errors.New("gorle: unsupported pixel data version")
)
