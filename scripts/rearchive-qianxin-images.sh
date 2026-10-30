#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/service"
# Repair corrupted raw-assets (HTML saved as .png): refetch live HTML then re-archive uploads.
# Example: ./scripts/rearchive-qianxin-images.sh -refetch-source -invalid-only
go run ./cmd/rearchive-assets/ -data ./data -url '%ti.qianxin.com%' "$@"
