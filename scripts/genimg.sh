#!/usr/bin/env bash
# genimg.sh — batch image generation via enowX (codebuddy/gemini-3.1-flash-image).
#
# Usage:
#   scripts/genimg.sh <manifest.json>
#
# Manifest JSON shape:
#   {
#     "model": "codebuddy/gemini-3.1-flash-image",   # optional
#     "size":  "1792x1024",                          # optional, default 1792x1024
#     "out_dir": "frontend/static/design-refs",     # optional
#     "items": [
#       { "name": "01-landing-hero", "prompt": "..." },
#       { "name": "02-landing-map",  "prompt": "..." }
#     ]
#   }
#
# Auth:
#   Set ENOWX_API_KEY env, OR script will read from
#   ~/.config/opencode/opencode.json -> provider.enowxai.options.apiKey
#
# Output:
#   <out_dir>/<name>.png   (only written on HTTP 200 + valid PNG header)
#   <out_dir>/.manifest.log  (one line per attempt: name | http | bytes | url_prefix | err)

set -u
set -o pipefail

CHANNEL="${CHANNEL:-genimg}"
log() { echo "[$CHANNEL] $*" >&2; }

if [ $# -lt 1 ]; then
  log "usage: $0 <manifest.json>"
  exit 2
fi
MANIFEST="$1"
[ -f "$MANIFEST" ] || { log "manifest not found: $MANIFEST"; exit 2; }

# ----- resolve API key -----
if [ -z "${ENOWX_API_KEY:-}" ]; then
  CONF="$HOME/.config/opencode/opencode.json"
  if [ -f "$CONF" ]; then
    ENOWX_API_KEY=$(python3 - "$CONF" <<'PY' 2>/dev/null
import json, sys
d = json.load(open(sys.argv[1]))
print((((d.get("provider") or {}).get("enowxai") or {}).get("options") or {}).get("apiKey") or "")
PY
    )
  fi
fi
if [ -z "${ENOWX_API_KEY:-}" ]; then
  log "ENOWX_API_KEY not set and not found in opencode config"
  exit 3
fi

BASE="${ENOWX_BASE_URL:-https://enowxai.ryanprayoga.dev/v1}"

# ----- parse manifest defaults -----
read -r MODEL SIZE OUT_DIR COUNT < <(python3 - "$MANIFEST" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
print(
    d.get("model","codebuddy/gemini-3.1-flash-image"),
    d.get("size","1792x1024"),
    d.get("out_dir","frontend/static/design-refs"),
    len(d.get("items",[])),
)
PY
)
mkdir -p "$OUT_DIR"
LOG="$OUT_DIR/.manifest.log"
: > "$LOG"

log "model=$MODEL size=$SIZE out_dir=$OUT_DIR items=$COUNT"

# ----- iterate items -----
i=0
ok=0
fail=0
while IFS=$'\t' read -r NAME PROMPT; do
  i=$((i+1))
  [ -z "$NAME" ] && continue
  OUT="$OUT_DIR/$NAME.png"
  if [ -f "$OUT" ] && [ "${FORCE:-0}" != "1" ]; then
    log "[$i/$COUNT] skip $NAME (exists, set FORCE=1 to regen)"
    echo "$NAME	skip	existing	-	-" >>"$LOG"
    continue
  fi

  REQ=$(python3 -c "
import json, sys
print(json.dumps({
    'model': sys.argv[1],
    'prompt': sys.argv[2],
    'n': 1,
    'size': sys.argv[3],
}))
" "$MODEL" "$PROMPT" "$SIZE")

  RESP_FILE=$(mktemp)
  HTTP=$(curl -s -m 180 -o "$RESP_FILE" -w '%{http_code}' \
    -H "Authorization: Bearer $ENOWX_API_KEY" \
    -H 'Content-Type: application/json' \
    -d "$REQ" \
    "$BASE/images/generations")

  if [ "$HTTP" != "200" ]; then
    ERR=$(head -c 240 "$RESP_FILE" | tr '\n\t' '  ')
    log "[$i/$COUNT] FAIL $NAME http=$HTTP err=$ERR"
    echo "$NAME	$HTTP	0	-	$ERR" >>"$LOG"
    fail=$((fail+1))
    rm -f "$RESP_FILE"
    continue
  fi

  URL=$(python3 -c "
import json,sys
try:
    d=json.load(open(sys.argv[1]))
    arr=d.get('data',[])
    if arr:
        print(arr[0].get('url') or arr[0].get('b64_json',''))
except Exception:
    pass
" "$RESP_FILE")
  rm -f "$RESP_FILE"

  if [ -z "$URL" ]; then
    log "[$i/$COUNT] FAIL $NAME no-url-in-response"
    echo "$NAME	200	0	-	no-url" >>"$LOG"
    fail=$((fail+1))
    continue
  fi

  curl -s -m 120 -o "$OUT" "$URL"
  BYTES=$(wc -c <"$OUT" | tr -d ' ')
  HEADER=$(python3 -c "
import sys
b=open(sys.argv[1],'rb').read(8)
print('PNG' if b==b'\\x89PNG\\r\\n\\x1a\\n' else 'BAD')
" "$OUT")

  if [ "$HEADER" != "PNG" ] || [ "$BYTES" -lt 1024 ]; then
    log "[$i/$COUNT] FAIL $NAME header=$HEADER bytes=$BYTES"
    echo "$NAME	200	$BYTES	${URL:0:80}	bad-png" >>"$LOG"
    fail=$((fail+1))
    rm -f "$OUT"
    continue
  fi

  log "[$i/$COUNT] OK   $NAME bytes=$BYTES"
  echo "$NAME	200	$BYTES	${URL:0:80}	-" >>"$LOG"
  ok=$((ok+1))
done < <(python3 - "$MANIFEST" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
for it in d.get("items", []):
    name = (it.get("name") or "").strip()
    prompt = (it.get("prompt") or "").replace("\t"," ").replace("\n"," ").strip()
    if name and prompt:
        print(f"{name}\t{prompt}")
PY
)

log "done ok=$ok fail=$fail total=$i log=$LOG"
[ "$fail" -eq 0 ]
