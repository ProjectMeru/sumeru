#!/usr/bin/env bash
# Fail when ContextWithBypass appears outside allowlisted definition/test files.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ALLOWLIST=(
  'core/orm/security_context.go'
  'core/orm/elevated_context.go'
  'core/orm/password_policy.go'
)

is_allowlisted() {
  local file="$1"
  for allowed in "${ALLOWLIST[@]}"; do
    if [[ "$file" == "$allowed" ]]; then
      return 0
    fi
  done
  [[ "$file" == test/* ]] && return 0
  return 1
}

violations=()
while IFS= read -r line; do
  file="${line%%:*}"
  if ! is_allowlisted "$file"; then
    violations+=("$line")
  fi
done < <(rg -n 'ContextWithBypass\(' --glob '*.go' || true)

if ((${#violations[@]} > 0)); then
  echo "ContextWithBypass used outside allowlist:" >&2
  printf '  %s\n' "${violations[@]}" >&2
  echo "Use orm.WithElevated instead (see core/orm/elevated_context.go)." >&2
  exit 1
fi

# Ban bypass on HTTP handler request context in core/server/web.
if rg -n 'ContextWithBypass\(r\.Context\(\)' core/server/web --glob '*.go' 2>/dev/null; then
  echo "ContextWithBypass must not wrap r.Context() in web handlers." >&2
  exit 1
fi

echo "security bypass check passed"
