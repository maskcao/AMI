#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/service"
go run ./cmd/rearchive-assets/ -data ./data -url '%ti.qianxin.com%' -dry-run "$@"
