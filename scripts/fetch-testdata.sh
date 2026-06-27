#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if ! python3 -c "import ljdata" 2>/dev/null; then
  echo "Installing pylibjpeg-data..."
  pip3 install "git+https://github.com/pydicom/pylibjpeg-data"
fi
if ! python3 -c "import pydicom" 2>/dev/null; then
  echo "Installing pydicom..."
  pip3 install pydicom
fi

python3 - <<'PY'
import json
import pathlib

from ljdata import get_indexed_datasets
from pydicom.encaps import generate_frames

uid = "1.2.840.10008.1.2.5"
index = get_indexed_datasets(uid)
out = pathlib.Path("testdata/dcm")
out.mkdir(parents=True, exist_ok=True)

names = [
    "OBXXXX1A_rle.dcm",
    "OBXXXX1A_rle_2frame.dcm",
    "SC_rgb_rle.dcm",
    "SC_rgb_rle_2frame.dcm",
    "MR_small_RLE.dcm",
    "emri_small_RLE.dcm",
    "SC_rgb_rle_16bit.dcm",
    "SC_rgb_rle_16bit_2frame.dcm",
    "rtdose_rle_1frame.dcm",
    "rtdose_rle.dcm",
    "SC_rgb_rle_32bit.dcm",
    "SC_rgb_rle_32bit_2frame.dcm",
]

for fname in names:
    ds = index[fname]["ds"]
    nr = int(getattr(ds, "NumberOfFrames", 1) or 1)
    frames = list(generate_frames(ds.PixelData, number_of_frames=nr))
    for fi, raw in enumerate(frames):
        if isinstance(raw, (list, tuple)):
            frame = b"".join(raw)
        else:
            frame = raw
        frame_path = out / (fname + ".frame")
        meta_path = out / (fname + ".json")
        if fi > 0:
            frame_path = out / (f"{fname}.f{fi}.frame")
            meta_path = out / (f"{fname}.f{fi}.json")
        frame_path.write_bytes(frame)
        meta = {
            "rows": int(ds.Rows),
            "columns": int(ds.Columns),
            "samples_per_pixel": int(ds.SamplesPerPixel),
            "bits_allocated": int(ds.BitsAllocated),
            "pixel_representation": int(getattr(ds, "PixelRepresentation", 0)),
            "number_of_frames": nr,
            "frame_index": fi,
        }
        meta_path.write_text(json.dumps(meta, indent=2))
        print(f"exported {frame_path.name}")
PY

echo "testdata ready under testdata/dcm"
