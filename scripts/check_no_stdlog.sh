#!/usr/bin/env bash
# Fail when core/ uses stdlib log or fmt.Print* for logging (use core/applog).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

violations=()
while IFS= read -r line; do
  violations+=("$line")
done < <(rg -n '\blog\.(Print|Fatal|Panic)|\bfmt\.Print' core --glob '*.go' --glob '!**/export_test.go' || true)

if ((${#violations[@]} > 0)); then
  echo "stdlib log/fmt.Print in core/ (use core/applog):" >&2
  printf '  %s\n' "${violations[@]}" >&2
  exit 1
fi

echo "no stdlog check passed"
