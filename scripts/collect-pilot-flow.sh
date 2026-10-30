#!/usr/bin/env bash
# Generic collect pilot: create/patch source -> list test -> article test -> unlock -> run.
# Requires: curl, python3, AMI API on AMI_API (default http://127.0.0.1:8080/api/v1).
#
# Required env:
#   SOURCE_NAME   Display name
#   CONFIG_JSON   Full config object as JSON string, or path to .json file
#
# Optional:
#   SOURCE_TYPE   default paginated_list
#   SOURCE_LABEL  default empty
#   AMI_API       default http://127.0.0.1:8080/api/v1
#   ARTICLE_URL   default: first sampleUrls from list test
#   LIST_MAX_PAGES  default 3
#   SKIP_RUN      set 1 to skip batch run after unlock
set -euo pipefail

API="${AMI_API:-http://127.0.0.1:8080/api/v1}"
SOURCE_NAME="${SOURCE_NAME:?SOURCE_NAME required}"
SOURCE_TYPE="${SOURCE_TYPE:-paginated_list}"
LIST_MAX_PAGES="${LIST_MAX_PAGES:-3}"

if [[ -z "${CONFIG_JSON:-}" ]]; then
  echo "CONFIG_JSON required (JSON string or file path)" >&2
  exit 1
fi
if [[ -f "$CONFIG_JSON" ]]; then
  CONFIG_JSON=$(cat "$CONFIG_JSON")
fi

curl -sf "${API%/api/v1}/health" >/dev/null

EXISTING=$(curl -sf "$API/settings/collect-sources" | SOURCE_NAME="$SOURCE_NAME" python3 -c "
import json,os,sys
name=os.environ['SOURCE_NAME']
for it in json.load(sys.stdin).get('items',[]):
    if it.get('name')==name:
        print(it['id']); break
")

if [[ -n "${EXISTING:-}" ]]; then
  ID="$EXISTING"
  curl -sf -X PATCH "$API/settings/collect-sources/$ID" \
    -H 'Content-Type: application/json' \
    -d "$(SOURCE_NAME="$SOURCE_NAME" CONFIG_JSON="$CONFIG_JSON" python3 -c "
import json,os
print(json.dumps({'name':os.environ['SOURCE_NAME'],'config':json.loads(os.environ['CONFIG_JSON'])}))
")" >/dev/null
  echo "Patched source $ID"
else
  ID=$(SOURCE_NAME="$SOURCE_NAME" SOURCE_TYPE="$SOURCE_TYPE" SOURCE_LABEL="${SOURCE_LABEL:-}" CONFIG_JSON="$CONFIG_JSON" python3 -c "
import json,os
body={
  'name':os.environ['SOURCE_NAME'],
  'type':os.environ['SOURCE_TYPE'],
  'enabled':False,
  'sourceLabel':os.environ.get('SOURCE_LABEL',''),
  'scheduleLabel':'manual',
  'config':json.loads(os.environ['CONFIG_JSON']),
}
print(json.dumps(body))
" | curl -sf -X POST "$API/settings/collect-sources" -H 'Content-Type: application/json' -d @- \
    | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])")
  echo "Created source $ID"
fi

LIST=$(curl -sf -X POST "$API/settings/collect-sources/$ID/test/list" \
  -H 'Content-Type: application/json' -d "{\"maxPages\":$LIST_MAX_PAGES}")
echo "$LIST" | python3 -m json.tool
echo "$LIST" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('passed') and (d.get('linkCount') or 0)>0 else 1)"

ARTICLE_URL="${ARTICLE_URL:-}"
if [[ -z "$ARTICLE_URL" ]]; then
  ARTICLE_URL=$(echo "$LIST" | python3 -c "import json,sys; u=json.load(sys.stdin).get('sampleUrls') or []; print(u[0] if u else '')")
fi
[[ -n "$ARTICLE_URL" ]] || { echo "No article URL" >&2; exit 1; }

ART=$(curl -sf -X POST "$API/settings/collect-sources/$ID/test/article" \
  -H 'Content-Type: application/json' \
  -d "{\"url\":\"$ARTICLE_URL\",\"waitPipeline\":true}")
echo "$ART" | python3 -m json.tool
echo "$ART" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('passed') else 1)"

curl -sf -X POST "$API/settings/collect-sources/$ID/unlock-batch" -d '{}' | python3 -m json.tool

if [[ "${SKIP_RUN:-0}" != "1" ]]; then
  curl -sf -X POST "$API/settings/collect-sources/$ID/run" -d '{}' | python3 -m json.tool
fi

echo "Pilot flow complete: $ID"
