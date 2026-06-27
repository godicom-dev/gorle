# gorle

Go DICOM **RLE Lossless** codec — **pure Go**, no CGO.

Aligned with [pylibjpeg-rle](https://github.com/pydicom/pylibjpeg-rle) for transfer syntax `1.2.840.10008.1.2.5`.

## Status

**Phase 1:** decode/encode frame API, PackBits 1-bit helpers, DICOM pixel data wrappers.

## Installation

```bash
go get github.com/godicom-dev/gorle
```

## Usage

### Decode one encapsulated frame

`DecodeFrame` returns **planar configuration 1** bytes (all R, then all G, then all B for RGB).  
Pass `rows * columns` as `nrPixels` (not the byte length).

```go
package main

import (
	"fmt"
	"log"

	"github.com/godicom-dev/gorle"
)

func main() {
	// `frame` is one item from DICOM encapsulated Pixel Data (OB/OW).
	var frame []byte

	rows, cols := 512, 512
	pixels, err := gorle.DecodeFrame(frame, rows*cols, 16, gorle.LittleEndian)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("decoded %d bytes (planar config 1)\n", len(pixels))
}
```

### Encode one frame

`EncodeFrame` expects **planar configuration 0** input (`R1,G1,B1,R2,G2,B2,…`).

```go
rows, cols, spp := 64, 64, 3
// PC0: interleaved samples per pixel
src := make([]byte, rows*cols*spp*2)

encoded, err := gorle.EncodeFrame(src, rows, cols, spp, 16, gorle.LittleEndian)
if err != nil {
	log.Fatal(err)
}
```

### DICOM pixel data helpers

`DecodePixelData` / `EncodePixelData` mirror pylibjpeg-rle `decode_pixel_data` / `encode_pixel_data`.

```go
// Decode (v2 returns raw bytes, like pylibjpeg Version.v2)
out, err := gorle.DecodePixelData(frame, gorle.PixelDataOptions{
	Version: gorle.PixelDataV2,
	FrameOptions: gorle.FrameOptions{
		Rows:          512,
		Columns:       512,
		BitsAllocated: 16,
		ByteOrder:     gorle.LittleEndian,
	},
})

// Encode
enc, err := gorle.EncodePixelData(pc0Pixels, gorle.PixelDataOptions{
	FrameOptions: gorle.FrameOptions{
		Rows:            512,
		Columns:         512,
		SamplesPerPixel: 1,
		BitsAllocated:   16,
		ByteOrder:       gorle.LittleEndian,
	},
})
```

For **1-bit** images, `DecodePixelData` with `PackBits: true` (v2 only) returns packed bits;  
`EncodePixelData` accepts either packed or unpacked 1-bit input.

### Low-level segment API

```go
offsets, err := gorle.ParseHeader(frame[:64])
seg, err := gorle.DecodeSegment(frame[offsets[0]:offsets[1]])
row, err := gorle.EncodeRow([]byte{1, 2, 3, 3, 3, 4})
```

## API

```go
func DecodeFrame(src []byte, nrPixels, bitsAllocated int, byteOrder ByteOrder) ([]byte, error)
func EncodeFrame(src []byte, rows, cols, spp, bitsAllocated int, byteOrder ByteOrder) ([]byte, error)
func DecodePixelData(src []byte, opts PixelDataOptions) ([]byte, error)
func EncodePixelData(src []byte, opts PixelDataOptions) ([]byte, error)
func ParseHeader(src []byte) ([]uint32, error)
func DecodeSegment(src []byte) ([]byte, error)
func EncodeSegment(src []byte, cols int) ([]byte, error)
func PackBits(src []byte, byteOrder ByteOrder) ([]byte, error)
func UnpackBits(src []byte, count int, byteOrder ByteOrder) ([]byte, error)
```

| Direction | Pixel layout |
|-----------|----------------|
| `EncodeFrame` / `EncodePixelData` input | Planar configuration **0** |
| `DecodeFrame` / `DecodePixelData` output | Planar configuration **1** |

Supported: `SamplesPerPixel` 1 or 3; `BitsAllocated` 1, 8, 16, 32, 64.

## Development

```bash
git clone https://github.com/godicom-dev/gorle.git
cd gorle
go test ./...
```

Optional cross-check against pylibjpeg-rle (Python tests skip if not installed):

```bash
pip install pylibjpeg-rle
go test -v ./...
```

## References

- [golibjpeg](https://github.com/godicom-dev/golibjpeg) — JPEG / JPEG-LS
- [goopenjpeg](https://github.com/godicom-dev/goopenjpeg) — JPEG 2000
- [pylibjpeg-rle](https://github.com/pydicom/pylibjpeg-rle) — behaviour reference
