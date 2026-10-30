#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/service"
go run ./cmd/backfill-disclosed/ -data ./data "$@"
