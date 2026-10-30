#!/usr/bin/env bash
# Backfill intel_report.title using collect_source title_extract rules (e.g. Qianxin Nuxt blog).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DATA_DIR="${AMI_DATA_DIR:-$ROOT/service/data}"
DRY_RUN="${DRY_RUN:-}"

ARGS=(-data "$DATA_DIR")
if [[ "${DRY_RUN}" == "1" || "${DRY_RUN}" == "true" ]]; then
  ARGS+=(-dry-run)
fi

cd "$ROOT/service"
exec go run ./cmd/backfill-titles "${ARGS[@]}"
