package gorle_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/godicom-dev/gorle"
)

func requirePythonRLE(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
	out, err := exec.Command("python3", "-c", "import rle.rle").CombinedOutput()
	if err != nil {
		t.Skipf("pylibjpeg-rle not installed: %s", out)
	}
}

func TestCrossPython8Bit(t *testing.T) {
	requirePythonRLE(t)
	src := []byte{10, 10, 10, 20, 20, 30, 30, 30, 30, 40, 50, 60, 60, 60, 60, 70, 70, 80, 80, 90}
	script := `
import sys
from rle.rle import encode_frame, decode_frame
src = bytes([10,10,10,20,20,30,30,30,30,40,50,60,60,60,60,70,70,80,80,90])
enc = encode_frame(src, 4, 5, 1, 8, '<')
sys.stdout.buffer.write(enc)
`
	out, err := exec.Command("python3", "-c", script).Output()
	if err != nil {
		t.Fatal(err)
	}
	dec, err := gorle.DecodeFrame(out, 20, 8, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(src, dec) {
		t.Fatal("go decode of python-encoded frame mismatch")
	}
	enc, err := gorle.EncodeFrame(src, 4, 5, 1, 8, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, enc) {
		t.Fatal("go encode mismatch vs python")
	}
}

func TestCrossPython16BitRGB(t *testing.T) {
	requirePythonRLE(t)
	src := []byte{
		0, 1, 2, 10, 11, 12, 20, 21, 22, 30, 31, 32, 40, 41, 42, 50, 51, 52,
		0, 2, 4, 10, 12, 14, 20, 22, 24, 30, 32, 34, 40, 42, 44, 50, 52, 54,
	}
	rows, cols, spp := 2, 3, 3
	enc, err := gorle.EncodeFrame(src, rows, cols, spp, 16, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	path := tmp + "/frame.bin"
	if err := os.WriteFile(path, enc, 0o644); err != nil {
		t.Fatal(err)
	}
	script := `
import sys
from rle.rle import decode_frame
enc = open(sys.argv[1], 'rb').read()
dec = decode_frame(enc, 6, 16, '<')
sys.stdout.buffer.write(dec)
`
	out, err := exec.Command("python3", "-c", script, path).Output()
	if err != nil {
		t.Fatal(err)
	}
	dec, err := gorle.DecodeFrame(enc, rows*cols, 16, gorle.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, dec) {
		t.Fatal("python decode != go decode for go-encoded RGB frame")
	}
}
