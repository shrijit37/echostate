#!/usr/bin/env bash
# Package the EchoState Chrome extension for distribution.
# Usage: ./scripts/package-extension.sh [version] [out-dir]
set -euo pipefail

EXT_DIR="extension"
OUT_DIR="${2:-dist}"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"

mkdir -p "$OUT_DIR"

zip_ext() {
    local srcdir="$1" out="$2"
    python3 -c "
import zipfile, os
z = zipfile.ZipFile('$out', 'w')
for f in ['manifest.json', 'content.js', 'style.css']:
    z.write(os.path.join('$srcdir', f), f)
z.close()
"
}

if [ -n "${1:-}" ]; then
    TMPDIR=$(mktemp -d)
    for f in manifest.json content.js style.css; do cp "$EXT_DIR/$f" "$TMPDIR/"; done
    python3 -c "import json; m=json.load(open('$TMPDIR/manifest.json')); m['version']='$VERSION'; json.dump(m,open('$TMPDIR/manifest.json','w'),indent=2)"
    zip_ext "$TMPDIR" "$OUT_DIR/echostate-extension-${VERSION}.zip"
    rm -rf "$TMPDIR"
else
    zip_ext "$EXT_DIR" "$OUT_DIR/echostate-extension-${VERSION}.zip"
fi

echo "Created $OUT_DIR/echostate-extension-${VERSION}.zip"
