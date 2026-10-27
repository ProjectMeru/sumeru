#!/usr/bin/env bash
# Addon Go code uses the SDK. Generated model aliases may import the kernel registry.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

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
done < <(find addons -name '*.go' -print)

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
echo "addon sdk check passed"
