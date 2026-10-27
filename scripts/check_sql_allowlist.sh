#!/usr/bin/env bash
# Fail when core/server/web uses orm.DB outside the infrastructure allowlist.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ALLOWED=(
  auth.go
  login.go
  rpc_json.go
  session_browser.go
  testexports.go
)

violations=()
while IFS= read -r line; do
  file="${line%%:*}"
  base="${file##*/}"
  allowed=false
  for a in "${ALLOWED[@]}"; do
    if [[ "$base" == "$a" ]]; then
      allowed=true
      break
    fi
  done
  if ! $allowed; then
    violations+=("$line")
  fi
done < <(rg -n 'orm\.DB\.(Query|Exec)' core/server/web --glob '*.go' || true)

if ((${#violations[@]} > 0)); then
  echo "orm.DB usage outside web SQL allowlist:" >&2
  printf '  %s\n' "${violations[@]}" >&2
  echo "See docs/orm-sql-policy.md — use ORM or add a reviewed allowlist entry." >&2
  exit 1
fi

echo "sql allowlist check passed"
