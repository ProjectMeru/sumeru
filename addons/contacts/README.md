# Contacts

Central address book application over **`core.partner`** from `base`.

## Models

| Model | Role |
|-------|------|
| `core.partner` | Extended via model inherit in `models/partner_extend.go` (`customer_rank`) |

## Layout (reference addon)

This module follows the standard addon layout documented in `core/module/addon_template/MODULE_STANDARD.txt`:

- `security/` — groups and ACL
- `views/actions.xml` — window action
- `views/core_partner_{form,list,kanban}_views.xml` — primary views
- `views/core_partner_form_inherit_views.xml` — view inherit demo (xpath adds `customer_rank`)
- `views/menus.xml` — menus (loaded last)

## Install

```bash
go run ./cmd/sumeru -- -c sumeru.conf -i contacts
```
