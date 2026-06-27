package gorle_test

import (
	"encoding/binary"
	"testing"

	"github.com/godicom-dev/gorle"
	"github.com/godicom-dev/gorle/internal/testdata"
)

func decodeDCM(t *testing.T, name string) ([]byte, testdata.DCMFrameMeta) {
	t.Helper()
	frame, meta := testdata.RequireDCMFrame(t, name)
	out, err := gorle.DecodePixelData(frame, gorle.PixelDataOptions{
		Version: gorle.PixelDataV1,
		FrameOptions: gorle.FrameOptions{
			Rows:            meta.Rows,
			Columns:         meta.Columns,
			SamplesPerPixel: meta.SamplesPerPixel,
			BitsAllocated:   meta.BitsAllocated,
			ByteOrder:       gorle.LittleEndian,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return out, meta
}

// RLE decode returns planar configuration 1 (RRR…GGG…BBB…).
func planarU8(pix []byte, rows, cols, y, x, c int) byte {
	return pix[c*rows*cols+y*cols+x]
}

func planarU16(pix []byte, rows, cols, y, x, c int) uint16 {
	off := (c*rows*cols + y*cols + x) * 2
	return binary.LittleEndian.Uint16(pix[off:])
}

func planarU32(pix []byte, rows, cols, y, x, c int) uint32 {
	off := (c*rows*cols + y*cols + x) * 4
	return binary.LittleEndian.Uint32(pix[off:])
}

func i16At(pix []byte, cols, y, x int) int16 {
	off := (y*cols + x) * 2
	return int16(binary.LittleEndian.Uint16(pix[off:]))
}

func TestReferenceDCM_U8_1s_1f(t *testing.T) {
	pix, meta := decodeDCM(t, "OBXXXX1A_rle.dcm")
	if meta.Rows != 600 || meta.Columns != 800 {
		t.Fatalf("unexpected dims: %+v", meta)
	}
	if len(pix) != 600*800 {
		t.Fatalf("len %d want %d", len(pix), 600*800)
	}
	if pix[0] != 244 {
		t.Fatalf("row0 min/max: got %d want 244", pix[0])
	}
	if pix[300*800+491] != 1 || pix[300*800+492] != 246 || pix[300*800+493] != 1 {
		t.Fatalf("arr[300,491:494]: got %d %d %d", pix[300*800+491], pix[300*800+492], pix[300*800+493])
	}
	last := pix[len(pix)-800:]
	for _, b := range last {
		if b != 0 {
			t.Fatalf("last row min/max: got %d want 0", b)
		}
	}
}

func TestReferenceDCM_U8_3s_1f(t *testing.T) {
	pix, meta := decodeDCM(t, "SC_rgb_rle.dcm")
	cases := [][3]byte{
		{255, 0, 0}, {255, 128, 128}, {0, 255, 0}, {128, 255, 128},
		{0, 0, 255}, {128, 128, 255}, {0, 0, 0}, {64, 64, 64},
		{192, 192, 192}, {255, 255, 255},
	}
	rows := []int{5, 15, 25, 35, 45, 55, 65, 75, 85, 95}
	for i, y := range rows {
		got := [3]byte{
			planarU8(pix, meta.Rows, meta.Columns, y, 50, 0),
			planarU8(pix, meta.Rows, meta.Columns, y, 50, 1),
			planarU8(pix, meta.Rows, meta.Columns, y, 50, 2),
		}
		if got != cases[i] {
			t.Fatalf("arr[%d,50,:]: got %v want %v", y, got, cases[i])
		}
	}
}

func TestReferenceDCM_I16_1s_1f(t *testing.T) {
	pix, meta := decodeDCM(t, "MR_small_RLE.dcm")
	if got := [3]int16{i16At(pix, meta.Columns, 0, 31), i16At(pix, meta.Columns, 0, 32), i16At(pix, meta.Columns, 0, 33)}; got != [3]int16{422, 319, 361} {
		t.Fatalf("arr[0,31:34]: got %v", got)
	}
	if got := [3]int16{i16At(pix, meta.Columns, 31, 0), i16At(pix, meta.Columns, 31, 1), i16At(pix, meta.Columns, 31, 2)}; got != [3]int16{366, 363, 322} {
		t.Fatalf("arr[31,:3]: got %v", got)
	}
	if got := [3]int16{i16At(pix, meta.Columns, 63, 61), i16At(pix, meta.Columns, 63, 62), i16At(pix, meta.Columns, 63, 63)}; got != [3]int16{1369, 1129, 862} {
		t.Fatalf("arr[-1,-3:]: got %v", got)
	}
}

func TestReferenceDCM_U16_1s_10f(t *testing.T) {
	pix, meta := decodeDCM(t, "emri_small_RLE.dcm")
	if meta.NumberOfFrames != 10 {
		t.Fatalf("frames: got %d want 10", meta.NumberOfFrames)
	}
	u16 := func(y, x int) uint16 {
		off := (y*64 + x) * 2
		return binary.LittleEndian.Uint16(pix[off:])
	}
	if got := [3]uint16{u16(0, 31), u16(0, 32), u16(0, 33)}; got != [3]uint16{206, 197, 159} {
		t.Fatalf("frame1 arr[0,31:34]: got %v", got)
	}
}

func TestReferenceDCM_U16_3s_1f(t *testing.T) {
	pix, meta := decodeDCM(t, "SC_rgb_rle_16bit.dcm")
	if got := [3]uint16{planarU16(pix, meta.Rows, meta.Columns, 5, 50, 0), planarU16(pix, meta.Rows, meta.Columns, 5, 50, 1), planarU16(pix, meta.Rows, meta.Columns, 5, 50, 2)}; got != [3]uint16{65535, 0, 0} {
		t.Fatalf("arr[5,50,:]: got %v", got)
	}
	if got := [3]uint16{planarU16(pix, meta.Rows, meta.Columns, 95, 50, 0), planarU16(pix, meta.Rows, meta.Columns, 95, 50, 1), planarU16(pix, meta.Rows, meta.Columns, 95, 50, 2)}; got != [3]uint16{65535, 65535, 65535} {
		t.Fatalf("arr[95,50,:]: got %v", got)
	}
}

func TestReferenceDCM_U32_1s_1f(t *testing.T) {
	pix, meta := decodeDCM(t, "rtdose_rle_1frame.dcm")
	u32 := func(y, x int) uint32 {
		off := (y*meta.Columns + x) * 4
		return binary.LittleEndian.Uint32(pix[off:])
	}
	if got := [3]uint32{u32(0, 0), u32(0, 1), u32(0, 2)}; got != [3]uint32{1249000, 1249000, 1250000} {
		t.Fatalf("arr[0,:3]: got %v", got)
	}
}

func TestReferenceDCM_U32_3s_1f(t *testing.T) {
	pix, meta := decodeDCM(t, "SC_rgb_rle_32bit.dcm")
	if got := [3]uint32{planarU32(pix, meta.Rows, meta.Columns, 5, 50, 0), planarU32(pix, meta.Rows, meta.Columns, 5, 50, 1), planarU32(pix, meta.Rows, meta.Columns, 5, 50, 2)}; got != [3]uint32{4294967295, 0, 0} {
		t.Fatalf("arr[5,50,:]: got %v", got)
	}
	if got := [3]uint32{planarU32(pix, meta.Rows, meta.Columns, 95, 50, 0), planarU32(pix, meta.Rows, meta.Columns, 95, 50, 1), planarU32(pix, meta.Rows, meta.Columns, 95, 50, 2)}; got != [3]uint32{4294967295, 4294967295, 4294967295} {
		t.Fatalf("arr[95,50,:]: got %v", got)
	}
}
