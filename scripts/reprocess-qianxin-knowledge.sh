#!/usr/bin/env bash
# Delete qianxin-linked attack_knowledge and re-run parse+extract pipeline.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="${AMI_DATA_DIR:-$ROOT/service/data}"
cd "$ROOT/service"
exec go run ./cmd/reprocess-qianxin-knowledge -data "$DATA_DIR" "$@"
