#!/usr/bin/env bash
# Renders each starter's preview.html to a small WebP (preview.webp) that the AI
# designer sends as a visual reference. Run from client/ with the app serving
# on APP_URL (default http://localhost:3117). Needs Google Chrome and cwebp.
#
#   bash scripts/capture-starter-previews.sh            # missing previews only
#   FORCE=1 bash scripts/capture-starter-previews.sh    # re-render all
set -euo pipefail

APP_URL="${APP_URL:-http://localhost:3117}"
CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
ROOT="public/assets/template-starters"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

command -v cwebp >/dev/null || { echo "cwebp not found (brew install webp)"; exit 1; }
[ -x "$CHROME" ] || { echo "Chrome not found at $CHROME"; exit 1; }

keys=$(python3 -c 'import json; print("\n".join(t["key"] for t in json.load(open("lib/template-starters/manifest.json"))))')
done=0
for key in $keys; do
  out="$ROOT/$key/preview.webp"
  [ -f "$out" ] && [ -z "${FORCE:-}" ] && continue
  # 600px is the email width; the first 1400px holds the masthead, hero and
  # first sections, which is what the designer needs to see.
  "$CHROME" --headless=new --disable-gpu --hide-scrollbars --virtual-time-budget=4000 \
    --window-size=600,1400 --screenshot="$TMP/$key.png" \
    "$APP_URL/assets/template-starters/$key/preview.html" >/dev/null 2>&1
  cwebp -quiet -resize 512 0 -q 72 "$TMP/$key.png" -o "$out"
  done=$((done + 1))
done
echo "Rendered $done previews."
