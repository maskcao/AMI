#!/usr/bin/env bash
# Wipe AMI local data (SQLite + artifacts). Optionally purge via API if server is up.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DATA="${AMI_DATA_DIR:-$ROOT/service/data}"
BASE="${AMI_BASE:-http://127.0.0.1:8080}"

if curl -sf -X POST "$BASE/api/v1/admin/purge-data" >/dev/null 2>&1; then
  echo "Purged via API: $BASE/api/v1/admin/purge-data"
  exit 0
fi

echo "Clearing data files in: $DATA"
rm -f "$DATA/ami.db" "$DATA/ami.db-wal" "$DATA/ami.db-shm"
rm -rf "$DATA/store"
mkdir -p "$DATA/store"
echo "Done. Start API: cd service && make run  (empty: AMI_SEED=0 go run ./cmd/ami)"
