import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchField, SwcSearchFilter, SwcWorkspacePayload } from "../../types/workspace.js";

export interface ControlPanelState {
  search: string;
  order?: string;
  offset: number;
  limit: number;
  selectedIds: Set<number>;
  filters: string[];
}

export function parseFilterCSV(raw?: string): string[] {
  return (raw ?? "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

export function toggleFilterName(active: string[], name: string): string[] {
  if (active.includes(name)) return active.filter((n) => n !== name);
  return [...active, name];
}

export interface ControlPanelOptions {
  payload: SwcWorkspacePayload;
  state: ControlPanelState;
  onPage: (offset: number) => void;
}

/** List toolbar secondary row: pagination only (shown when multiple pages). */
export function renderControlPanel(opts: ControlPanelOptions): TemplateResult {
  const { payload, state, onPage } = opts;
  const rows = payload.records ?? [];
  const total = payload.listTotal ?? rows.length;
  const page = Math.floor(state.offset / state.limit) + 1;
  const pageCount = Math.max(1, Math.ceil(total / state.limit));
  const showPager = pageCount > 1 || state.offset > 0;

  if (!showPager) return html``;

  return html`
    <div class="sum-list-control sum-list-control--secondary">
      <div class="sum-list-pager">
        <button
          type="button"
          class="sum-btn sum-btn--ghost"
          disabled=${state.offset <= 0 ? "disabled" : undefined}
          @click=${() => onPage(Math.max(0, state.offset - state.limit))}
        >
          Prev
        </button>
        <span>${page} / ${pageCount}</span>
        <button
          type="button"
          class="sum-btn sum-btn--ghost"
          disabled=${state.offset + state.limit >= total ? "disabled" : undefined}
          @click=${() => onPage(state.offset + state.limit)}
        >
          Next
        </button>
      </div>
    </div>
  `;
}

export function renderSearchFilters(opts: {
  filters: SwcSearchFilter[];
  active: string[];
  onToggle: (name: string) => void;
}): TemplateResult {
  const domainFilters = opts.filters.filter((f) => f.domain || !f.groupBy);
  const groupFilters = opts.filters.filter((f) => f.groupBy);
  if (domainFilters.length === 0 && groupFilters.length === 0) return html``;
  return html`
    <div class="sum-search-filters">
      ${domainFilters.map((f) => {
        const on = opts.active.includes(f.name);
        return html`<button
          type="button"
          class=${on ? "sum-search-chip sum-search-chip--active" : "sum-search-chip"}
          @click=${() => opts.onToggle(f.name)}
        >
          ${f.string || f.name}
        </button>`;
      })}
      ${groupFilters.length
        ? html`<span class="sum-search-filters-label">Group</span>${groupFilters.map((f) => {
            const on = opts.active.includes(f.name);
            return html`<button
              type="button"
              class=${on ? "sum-search-chip sum-search-chip--active" : "sum-search-chip"}
              @click=${() => opts.onToggle(f.name)}
            >
              ${f.string || f.name}
            </button>`;
          })}`
        : ""}
    </div>
  `;
}

export function renderSortHeader(
  field: SwcArchField,
  currentSort: string,
  onSort: (fieldName: string) => void,
): TemplateResult {
  const name = field.name;
  const active = currentSort === name || currentSort === `-${name}`;
  const desc = currentSort === `-${name}`;
  const marker = active ? (desc ? " ↓" : " ↑") : "";
  return html`<th
    class=${active ? "sum-list-th sum-list-th--sort" : "sum-list-th"}
    @click=${() => onSort(name)}
  >
    ${field.string ?? field.name}${marker}
  </th>`;
}

export function renderRowCheckbox(
  id: number,
  selected: boolean,
  onToggle: (id: number, checked: boolean) => void,
): TemplateResult {
  return html`<td class="sum-list-select-cell" @click=${(ev: Event) => ev.stopPropagation()}>
    <input
      type="checkbox"
      checked=${selected ? "checked" : undefined}
      @change=${(ev: Event) => onToggle(id, (ev.target as HTMLInputElement).checked)}
    />
  </td>`;
}

export function renderSelectAllHeader(
  allSelected: boolean,
  onToggleAll: (checked: boolean) => void,
): TemplateResult {
  return html`<th class="sum-list-select-head">
    <input
      type="checkbox"
      title="Select all"
      checked=${allSelected ? "checked" : undefined}
      @change=${(ev: Event) => onToggleAll((ev.target as HTMLInputElement).checked)}
    />
  </th>`;
}
