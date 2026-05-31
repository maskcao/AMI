#!/usr/bin/env bash
# Smoke test AMI API (expects server on AMI_ADDR, default :8080)
set -euo pipefail
BASE="${AMI_BASE:-http://127.0.0.1:8080}"

echo "== health =="
curl -sf "$BASE/health" | head -c 80
echo

echo "== intel list =="
curl -sf "$BASE/api/v1/intel" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['total']>=1, d; print('total', d['total'])"

echo "== knowledge list =="
curl -sf "$BASE/api/v1/knowledge" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['total']>=1; print('total', d['total'])"

echo "== knowledge 1001 =="
curl -sf "$BASE/api/v1/knowledge/1001" | python3 -c "import sys,json; d=json.load(sys.stdin); assert len(d.get('linkedIntel',[]))>=1; print('linked', len(d['linkedIntel']))"

echo "== intel 4203 original =="
code=$(curl -s -o /tmp/ami-orig.html -w '%{http_code}' "$BASE/api/v1/intel/4203/original")
test "$code" = "200" && test "$(wc -c </tmp/ami-orig.html)" -gt 20
echo "original bytes $(wc -c </tmp/ami-orig.html)"

echo "== patch + publish =="
curl -sf -X PATCH "$BASE/api/v1/knowledge/1001" -H 'Content-Type: application/json' -d '{"summary":"smoke"}' >/dev/null
curl -sf -X POST "$BASE/api/v1/knowledge/1001/publish" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['status']=='published'; print('published')"

echo "== vite proxy (optional) =="
if curl -sf "http://127.0.0.1:5173/api/v1/intel" >/dev/null 2>&1; then
  echo "proxy OK"
else
  echo "proxy skip (webui not running)"
fi

echo "== analytics bundle =="
curl -sf "$BASE/api/v1/analytics/bundle" | python3 -c "import sys,json; d=json.load(sys.stdin); assert 'dashboard' in d and len(d.get('cve',[]))>=1; print('cve',len(d['cve']),'knowledge',d['dashboard']['snapshot']['knowledgeCount'])"

echo "== runtime logs =="
curl -sf "$BASE/api/v1/datacenter/runtime-logs" | python3 -c "import sys,json; d=json.load(sys.stdin); assert 'items' in d; print('logs', len(d['items']), 'total', d.get('total',0))"

echo "== settings =="
curl -sf "$BASE/api/v1/settings/collect-sources" | python3 -c "import sys,json; d=json.load(sys.stdin); assert len(d['items'])>=3; print('sources',len(d['items']))"

echo "== settings patch + collect run =="
curl -sf -X PATCH "$BASE/api/v1/settings/model-providers/mp-1" -H 'Content-Type: application/json' \
  -d '{"enabled":true}' >/dev/null
curl -sf -X POST "$BASE/api/v1/settings/collect-sources/cs-2/run" | python3 -c "import sys,json; d=json.load(sys.stdin); print('run', d.get('imported'), d.get('skipped'), d.get('failed',0))"

echo "ALL SMOKE PASSED"
