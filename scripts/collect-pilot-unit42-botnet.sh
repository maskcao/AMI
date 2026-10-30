#!/usr/bin/env bash
# Pilot + batch workflow for Unit42 botnet tag (WordPress AJAX pagination).
set -euo pipefail

API="${AMI_API:-http://127.0.0.1:8080/api/v1}"
SOURCE_NAME="${SOURCE_NAME:-Unit42 botnet tag}"

CONFIG='{
  "list_url": "https://unit42.paloaltonetworks.com/tag/botnet/",
  "article_base": "https://unit42.paloaltonetworks.com",
  "link_extract": {
    "must_contain": "unit42.paloaltonetworks.com",
    "exclude_contains": [
      "/tag/", "/page/", "/feed/", "/category/", "/wp-",
      "/tools/", "/atoms/", "/about-unit", "/unit-42-threat", "/ja/"
    ]
  },
  "pagination": {
    "type": "wordpress_ajax",
    "max_pages": 5,
    "stop_on_empty_list": true
  },
  "max_items_per_run": 50,
  "delay_ms": 1500,
  "timeout_sec": 45
}'

echo "==> Health"
curl -sf "${API%/api/v1}/health" >/dev/null

echo "==> Create or reuse collect source"
EXISTING=$(curl -sf "$API/settings/collect-sources" | python3 -c "
import json,sys
name=sys.argv[1]
for it in json.load(sys.stdin).get('items',[]):
    if it.get('name')==name:
        print(it['id']); break
" "$SOURCE_NAME" || true)

if [[ -n "${EXISTING:-}" ]]; then
  ID="$EXISTING"
  echo "    patch existing $ID"
  curl -sf -X PATCH "$API/settings/collect-sources/$ID" \
    -H 'Content-Type: application/json' \
    -d "{\"config\":$CONFIG}" >/dev/null
else
  BODY=$(python3 -c "import json; print(json.dumps({
    'name': '$SOURCE_NAME',
    'type': 'paginated_list',
    'enabled': False,
    'sourceLabel': 'Unit42',
    'scheduleLabel': 'manual',
    'config': json.loads('''$CONFIG'''),
  }))")
  ID=$(curl -sf -X POST "$API/settings/collect-sources" \
    -H 'Content-Type: application/json' \
    -d "$BODY" | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])")
  echo "    created $ID"
fi

echo "==> ② List pilot (wordpress_ajax pages)"
LIST=$(curl -sf -X POST "$API/settings/collect-sources/$ID/test/list" \
  -H 'Content-Type: application/json' \
  -d '{"maxPages":3}')
echo "$LIST" | python3 -m json.tool
LINK_COUNT=$(echo "$LIST" | python3 -c "import json,sys; print(json.load(sys.stdin).get('linkCount',0))")
if [[ "$LINK_COUNT" -lt 1 ]]; then
  echo "List pilot failed: no links" >&2
  exit 1
fi
SAMPLE_URL=$(echo "$LIST" | python3 -c "import json,sys; u=json.load(sys.stdin).get('sampleUrls') or []; print(u[0] if u else '')")
if [[ -z "$SAMPLE_URL" ]]; then
  echo "No sample URL for article pilot" >&2
  exit 1
fi

echo "==> ① Article pilot: $SAMPLE_URL"
ART=$(curl -sf -X POST "$API/settings/collect-sources/$ID/test/article" \
  -H 'Content-Type: application/json' \
  -d "{\"url\":\"$SAMPLE_URL\",\"waitPipeline\":true}")
echo "$ART" | python3 -m json.tool
PASSED=$(echo "$ART" | python3 -c "import json,sys; d=json.load(sys.stdin); print('1' if d.get('passed') else '0')")
if [[ "$PASSED" != "1" ]]; then
  echo "Article pilot quality gate did not pass" >&2
  exit 1
fi

echo "==> ③ Unlock batch"
UNLOCK=$(curl -s -X POST "$API/settings/collect-sources/$ID/unlock-batch" \
  -H 'Content-Type: application/json' -d '{}')
echo "$UNLOCK" | python3 -m json.tool
echo "$UNLOCK" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('ok') else 1)"

echo "==> Batch run"
RUN=$(curl -sf -X POST "$API/settings/collect-sources/$ID/run" -H 'Content-Type: application/json' -d '{}')
echo "$RUN" | python3 -m json.tool

echo "==> Done source id=$ID"
