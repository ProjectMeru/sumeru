# SWC workspace views

Sumeru workspace views are TypeScript components registered in `main.ts`. View XML is parsed on the server (`swcmeta/arch.go`) into `SwcWorkspacePayload.arch` JSON; SWC does not parse XML at runtime.

## Shared modules

| Module | Responsibility |
|--------|----------------|
| `collection-layout.ts` | `sum-collection-view` shell wrapper + control bar slot |
| `collection-view.ts` | Base class for collection views; `renderShell()` |
| `collection-bar-host.ts` | Search, New, filters, group-by, favorites, actions |
| `collection-navigation.ts` | `openWorkspaceRecord()` — open a row in form view |
| `list-table.ts` | `renderArchListTable()` — simple arch-driven tables |
| `arch-fields.ts` | Field visibility, list/kanban/graph/pivot arch rules |
| `field-display.ts` | `formatFieldValue()`, `recordDisplayLabel()` |
| `form-chrome.ts` | Form toolbar (Save/Edit, header fields, reports) |

**Simple table vs ListView:** Hierarchy and Activity use `renderArchListTable()`. ListView keeps its own table (sections, checkboxes, sort, bulk delete) and does not share the simple table helper.

## Collection views

List, kanban, graph, pivot, calendar, gantt, map, cohort, hierarchy, and activity extend `CollectionView`:

- **Chrome:** `CollectionBarHost` owns search, New, filters, group-by, favorites, and actions.
- **Shell:** `renderCollectionShell()` / `CollectionView.renderShell()` wraps the control bar and view body in `sum-collection-view sum-{type}-view`.
- **Arch fields:** `arch-fields.ts` defines shared rules for visible columns, kanban card fields, graph axes, and pivot groups.

Each `*View.ts` file should focus on body rendering and view-specific interaction (drag-drop, Chart.js, etc.).

## Form view

`FormView` is standalone (not a `CollectionView`). Toolbar chrome lives in `form-chrome.ts` (Save/Cancel/Edit, header status fields, report actions). Sheet layout is in `form-sheet.ts`; field widgets and lightbox in `form-interactions.ts`.

## Breadcrumbs and tabs

Breadcrumbs and view switcher tabs stay server-rendered in `base.html` and are synced client-side by `breadcrumb-sync.ts` and `view-tab-sync.ts`.

## Kanban XML convention

Kanban views declare fields directly on `<view type="kanban">`:

- Visible fields render on cards.
- `invisible="1"` loads data-only fields (color, gender, group-by field).
- Kanban options (`default_group_by`, `columns_per_row`) map to `arch.kanban`.

Legacy `<templates>` blocks are ignored by SWC and should not be used.
