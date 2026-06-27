package gorle_test

import (
	"bytes"
	"testing"

	"github.com/godicom-dev/gorle"
)

func TestRoundTrip8BitMono(t *testing.T) {
	rows, cols := 4, 5
	src := []byte{
		10, 10, 10, 20, 20,
		30, 30, 30, 30, 40,
		50, 60, 60, 60, 60,
		70, 70, 80, 80, 90,
	}
	enc, err := gorle.EncodeFrame(src, rows, cols, 1, 8, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := gorle.DecodeFrame(enc, rows*cols, 8, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(src, dec) {
		t.Fatalf("round-trip mismatch:\nenc=% x\ndec=% x", enc, dec)
	}
}

func TestRoundTrip16BitRGB(t *testing.T) {
	rows, cols, spp := 2, 3, 3
	// Planar configuration 0: R,G,B per pixel.
	src := []byte{
		0, 1, 2, 10, 11, 12, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, 52,
		0, 2, 4, 10, 12, 14, 20, 22, 24, 30, 32, 34, 40, 42, 44, 50, 52, 54,
	}
	enc, err := gorle.EncodeFrame(src, rows, cols, spp, 16, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := gorle.DecodeFrame(enc, rows*cols, 16, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	got := planar1ToPlanar0(dec, rows, cols, 2, spp)
	if !bytes.Equal(src, got) {
		t.Fatal("16-bit RGB round-trip mismatch")
	}
}

func planar1ToPlanar0(frame []byte, rows, cols, bytesPerSample, spp int) []byte {
	pixels := rows * cols
	out := make([]byte, len(frame))
	for p := 0; p < pixels; p++ {
		for s := 0; s < spp; s++ {
			for b := 0; b < bytesPerSample; b++ {
				srcOff := s*pixels*bytesPerSample + p*bytesPerSample + b
				dstOff := p*spp*bytesPerSample + s*bytesPerSample + b
				out[dstOff] = frame[srcOff]
			}
		}
	}
	return out
}

func TestDecodePixelDataV2PackBits(t *testing.T) {
	rows, cols := 2, 5
	unpacked := []byte{0, 1, 0, 0, 1, 1, 0, 0, 0, 1}
	enc, err := gorle.EncodeFrame(unpacked, rows, cols, 1, 1, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	out, err := gorle.DecodePixelData(enc, gorle.PixelDataOptions{
		Version: gorle.PixelDataV2,
		FrameOptions: gorle.FrameOptions{
			Rows:          rows,
			Columns:       cols,
			BitsAllocated: 1,
			ByteOrder:     gorle.LittleEndian,
		},
		PackBits: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want, err := gorle.PackBits(unpacked, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, out) {
		t.Fatalf("pack_bits decode mismatch: got %x want %x", out, want)
	}
}

func TestParseHeader(t *testing.T) {
	frame, err := gorle.EncodeFrame([]byte{1, 2, 3, 4}, 2, 2, 1, 8, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	offsets, err := gorle.ParseHeader(frame[:64])
	if err != nil {
		t.Fatal(err)
	}
	if offsets[0] != 64 {
		t.Fatalf("first segment offset = %d, want 64", offsets[0])
	}
}

func TestEncodePixelDataRejectsTooManySegments(t *testing.T) {
	rows, cols, spp := 1, 1, 3
	src := make([]byte, spp*8)
	_, err := gorle.EncodePixelData(src, gorle.PixelDataOptions{
		FrameOptions: gorle.FrameOptions{
			Rows:            rows,
			Columns:         cols,
			SamplesPerPixel: spp,
			BitsAllocated:   64,
			ByteOrder:       gorle.LittleEndian,
		},
	})
	if err != gorle.ErrTooManySegments {
		t.Fatalf("got %v, want ErrTooManySegments", err)
	}
}

func TestPackUnpackBits(t *testing.T) {
	src := []byte{0, 1, 0, 0, 1, 1, 0, 0, 0, 1}
	packed, err := gorle.PackBits(src, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	unpacked, err := gorle.UnpackBits(packed, len(src), gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(src, unpacked) {
		t.Fatalf("pack/unpack mismatch: %v vs %v", unpacked, src)
	}
}
