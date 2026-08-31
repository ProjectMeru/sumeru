# Sumeru AI

Optional AI assistant hooks for the web shell and partner forms.

## Opt-in import

This module sets **`auto_import: false`** in `manifest.json`, so it is **not** included in generated `zimports.go` by default. To enable:

1. Install the module: `go run ./cmd/sumeru -- -c sumeru.conf -i sumeru_ai`
2. Or add a blank import in your runner: `_ "sumeru/addons/sumeru_ai"`

## Hooks (`hooks.go`)

- ORM search interceptor (placeholder for natural-language domain translation)
- Shell FAB button for AI assistant
- Notebook tab on `core.partner` forms

## Assets

- `static/css/theme-overrides.css` — optional theme tokens
