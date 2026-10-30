#!/usr/bin/env bash
# Qianxin WeChat album — weekly advanced threat digest + source article import.
set -euo pipefail
API="${AMI_API:-http://127.0.0.1:8080/api/v1}"
SOURCE_NAME="${SOURCE_NAME:-奇安信微信-每周高级威胁情报}"

LIST_URL='https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzI2MDc2MDA4OA==&action=getalbum&album_id=1533540987225096195&scene=126'

CONFIG=$(cat <<'EOF'
{
  "list_url": "https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzI2MDc2MDA4OA==&action=getalbum&album_id=1533540987225096195&scene=126",
  "article_base": "https://mp.weixin.qq.com",
  "link_extract": {
    "must_contain": "mp.weixin.qq.com",
    "title_must_contain": "每周高级威胁情报解读"
  },
  "pagination": {
    "type": "wechat_album",
    "max_pages": 20,
    "scroll_steps": 20
  },
  "digest_parser": "wechat_weekly_threat",
  "import_weekly_digest": false,
  "digest_source_max": 0,
  "max_items_per_run": 5000,
  "delay_ms": 2000,
  "timeout_sec": 60
}
EOF
)

echo "== create/update collect source =="
ID=$(curl -sf "$API/settings/collect-sources" | python3 -c "
import json,sys,os
name=os.environ.get('SOURCE_NAME','')
data=json.load(sys.stdin)
items=data if isinstance(data,list) else data.get('items',[])
for s in items:
    if s.get('name')==name:
        print(s['id']); sys.exit(0)
print('')
" SOURCE_NAME="$SOURCE_NAME")

if [[ -z "$ID" ]]; then
  ID=$(curl -sf -X POST "$API/settings/collect-sources" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"$SOURCE_NAME\",\"type\":\"wechat_feed\",\"sourceLabel\":\"wechat\",\"config\":$CONFIG}" \
    | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])")
else
  curl -sf -X PATCH "$API/settings/collect-sources/$ID" \
    -H 'Content-Type: application/json' \
    -d "{\"config\":$CONFIG}" >/dev/null
fi
echo "source id=$ID"

echo "== test/list (album pagination) =="
LIST_JSON=$(curl -sf -X POST "$API/settings/collect-sources/$ID/test/list" \
  -H 'Content-Type: application/json' -d '{"maxPages":5}')
echo "$LIST_JSON" | python3 -m json.tool
LINK_COUNT=$(echo "$LIST_JSON" | python3 -c "import json,sys; print(json.load(sys.stdin).get('linkCount',0))")
if [[ "$LINK_COUNT" -lt 1 ]]; then
  echo "FAIL: linkCount=0" >&2
  exit 1
fi
SAMPLE=$(echo "$LIST_JSON" | python3 -c "import json,sys; print(json.load(sys.stdin)['sampleUrls'][0])")

echo "== test/article on first weekly digest =="
curl -sf -X POST "$API/settings/collect-sources/$ID/test/article" \
  -H 'Content-Type: application/json' \
  -d "{\"url\":\"$SAMPLE\",\"waitPipeline\":true}" | python3 -m json.tool

echo "== unlock + run (imports 情报来源 articles; cap max_items_per_run) =="
curl -sf -X POST "$API/settings/collect-sources/$ID/unlock-batch" -d '{}' >/dev/null
curl -sf -X POST "$API/settings/collect-sources/$ID/run" -d '{}' | python3 -m json.tool
echo "OK"
