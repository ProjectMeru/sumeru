#!/usr/bin/env bash
# Go *_test.go must live under test/ only; SWC *.test.ts must not live under core/swc/src/.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

violations=()
while IFS= read -r path; do
  [[ -n "$path" ]] && violations+=("$path")
done < <(find . \( -path ./test -o -path ./.git \) -prune -o -name '*_test.go' -print 2>/dev/null)

if ((${#violations[@]} > 0)); then
  echo "Go test files outside test/ (move to test/ mirroring package path):" >&2
  printf '  %s\n' "${violations[@]}" >&2
  exit 1
fi

swc_violations=()
while IFS= read -r path; do
  [[ -n "$path" ]] && swc_violations+=("$path")
done < <(find core/swc/src -name '*.test.ts' -print 2>/dev/null || true)

if ((${#swc_violations[@]} > 0)); then
  echo "SWC test files under core/swc/src/ (use core/swc/tests/):" >&2
  printf '  %s\n' "${swc_violations[@]}" >&2
  exit 1
fi

echo "test layout check passed"
