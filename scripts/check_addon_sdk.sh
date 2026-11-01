#!/usr/bin/env bash
# Addon Go code uses the SDK. Generated model aliases may import the kernel registry.
# With no argument, scans sumeru/addons. Pass an addons directory to scan another tree.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
scan="${1:-$ROOT/addons}"

if [[ ! -d "$scan" ]]; then
  echo "addon sdk check: directory not found: $scan" >&2
  exit 1
fi

fail=0
while IFS= read -r path; do
  base="$(basename "$path")"
  case "$base" in
    zrefs.go|zmodels.go) continue ;;
  esac
  if grep -E -q '"database/sql"|sumeru/core/orm' "$path"; then
    echo "addon imports kernel data packages (use sumeru/core/sdk): $path" >&2
    fail=1
  fi
  if grep -E -q 'QueryContext|ExecContext|QueryRowContext' "$path"; then
    echo "addon runs SQL (use sumeru/core/sdk): $path" >&2
    fail=1
  fi
  if grep -E -q '`[^`]*(SELECT|INSERT|UPDATE|DELETE)[[:space:]]' "$path"; then
    echo "addon contains raw SQL (use sumeru/core/sdk): $path" >&2
    fail=1
  fi
done < <(find "$scan" -name '*.go' -print)

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
echo "addon sdk check passed"
