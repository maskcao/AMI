#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="${AMI_DATA_DIR:-$ROOT/service/data}"
cd "$ROOT/service"
exec go run ./cmd/repair-knowledge -data "$DATA_DIR" "$@"
