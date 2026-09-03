#!/usr/bin/env bash
# Enforce minimum Go statement coverage from go tool cover output.
# Override threshold: GO_COVERAGE_MIN=42 bash scripts/check_go_coverage.sh coverage.out
set -euo pipefail

PROFILE="${1:-coverage.out}"
MIN="${GO_COVERAGE_MIN:-90}"

if [[ ! -f "$PROFILE" ]]; then
  echo "coverage profile not found: $PROFILE" >&2
  exit 1
fi

TOTAL=$(go tool cover -func="$PROFILE" | awk '/^total:/ { gsub(/%/,"",$3); print $3; exit }')
if [[ -z "$TOTAL" ]]; then
  echo "could not parse total coverage from $PROFILE" >&2
  exit 1
fi

awk -v total="$TOTAL" -v min="$MIN" 'BEGIN {
  if (total + 0 < min + 0) {
    printf "FAIL: total statement coverage %.1f%% is below minimum %.1f%%\n", total, min
    exit 1
  }
  printf "OK: total statement coverage %.1f%% (min %.1f%%)\n", total, min
}'
