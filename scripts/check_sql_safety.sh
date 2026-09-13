#!/usr/bin/env bash
# Fail on ad-hoc SQL string building in web handlers (use ORM / quoted identifiers).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

violations=()
while IFS= read -r line; do
  violations+=("$line")
done < <(rg -n 'fmt\.Sprintf\("(SELECT|INSERT|UPDATE|DELETE)' core/server/web --glob '*.go' || true)

if ((${#violations[@]} > 0)); then
  echo "raw SQL fmt.Sprintf in core/server/web:" >&2
  printf '  %s\n' "${violations[@]}" >&2
  echo "Use orm helpers and MustQuotedTableName instead." >&2
  exit 1
fi

echo "sql safety check passed"
