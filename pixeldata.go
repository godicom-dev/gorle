package gorle

// DecodePixelData decodes a single RLE Lossless frame (pylibjpeg-rle decode_pixel_data).
func DecodePixelData(src []byte, opts PixelDataOptions) ([]byte, error) {
	if opts.Version == 0 {
		opts.Version = PixelDataV1
	}
	if opts.ByteOrder == "" {
		opts.ByteOrder = LittleEndian
	}

	rows := opts.Rows
	cols := opts.Columns
	bits := opts.BitsAllocated
	if rows == 0 || cols == 0 || bits == 0 {
		return nil, ErrMissingPixelDataArgs
	}

	frame, err := DecodeFrame(src, rows*cols, bits, opts.ByteOrder)
	if err != nil {
		return nil, err
	}

	switch opts.Version {
	case PixelDataV1:
		return frame, nil
	case PixelDataV2:
		if bits == 1 && opts.PackBits {
			return PackBits(frame, opts.ByteOrder)
		}
		return frame, nil
	default:
		return nil, ErrUnsupportedPixelVersion
	}
}

// EncodePixelData encodes a single frame using DICOM RLE (pylibjpeg-rle encode_pixel_data).
func EncodePixelData(src []byte, opts PixelDataOptions) ([]byte, error) {
	if opts.ByteOrder == "" {
		if opts.BitsAllocated <= 8 {
			opts.ByteOrder = LittleEndian
		}
	}

	rows := opts.Rows
	cols := opts.Columns
	spp := opts.SamplesPerPixel
	bits := opts.BitsAllocated
	if rows == 0 || cols == 0 || bits == 0 || spp == 0 {
		return nil, ErrMissingPixelDataArgs
	}

	if spp != 1 && spp != 3 {
		return nil, ErrInvalidSamplesPerPixel
	}
	if bits != 1 && bits != 8 && bits != 16 && bits != 32 && bits != 64 {
		return nil, ErrInvalidBitsAllocated
	}
	if bits != 1 && (bits/8)*spp > 15 {
		return nil, ErrTooManySegments
	}
	if bits > 8 && opts.ByteOrder != LittleEndian && opts.ByteOrder != BigEndian {
		return nil, ErrInvalidByteOrder
	}

	data := src
	if bits == 1 {
		packedLen := Packed1BitLength(rows, cols)
		if len(src) == packedLen {
			unpacked, err := UnpackBits(src, rows*cols, opts.ByteOrder)
			if err != nil {
				return nil, err
			}
			data = unpacked
		}
	}

	expected := UnpackedFrameLength(rows, cols, spp, bits)
	if len(data) != expected {
		return nil, ErrInvalidParameters
	}

	return EncodeFrame(data, rows, cols, spp, bits, opts.ByteOrder)
}
