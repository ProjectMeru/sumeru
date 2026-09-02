import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchField, SwcSearchFilter, SwcWorkspacePayload } from "../../types/workspace.js";
import { inputValueFromEvent } from "../../widgets/field-events.js";
import { buildReportActionEntries } from "./view-toolbar.js";
import {
  filterOperatorsForField,
  presetDomainFilters,
  presetGroupByFilters,
  type CollectionQuery,
} from "./collection-query.js";

export function renderPopoverCheckmark(active: boolean): TemplateResult {
  return html`<span class="sum-popover-check" aria-hidden="true">${active ? "✓" : ""}</span>`;
}

export function renderPopoverItem(label: string, active: boolean, onClick: () => void): TemplateResult {
  return html`<li>
    <button
      type="button"
      class=${active ? "sum-popover-item sum-popover-item--active" : "sum-popover-item"}
      @click=${onClick}
    >
      ${renderPopoverCheckmark(active)}
      <span class="sum-popover-item-label">${label}</span>
    </button>
  </li>`;
}

export interface FiltersPopoverState {
  query: CollectionQuery;
  customField: string;
  customOp: string;
  customValue: string;
  domainPresets: SwcSearchFilter[];
  filterFields: SwcArchField[];
}

export interface FiltersPopoverCallbacks {
  onTogglePreset: (name: string) => void;
  onCustomFieldChange: (field: string, defaultOp: string) => void;
  onCustomOpChange: (op: string) => void;
  onCustomValueInput: (value: string) => void;
  onApplyCustom: () => void;
}

export function renderFiltersPopover(
  state: FiltersPopoverState,
  callbacks: FiltersPopoverCallbacks,
): TemplateResult {
  const { query, customField, customOp, customValue, domainPresets, filterFields } = state;
  const field = customField || filterFields[0]?.name || "";
  const operators = filterOperatorsForField(field, filterFields);

  return html`
    <div class="sum-popover sum-popover--filters" @click=${(e: Event) => e.stopPropagation()}>
      <h3 class="sum-popover-heading">Filters</h3>
      <ul class="sum-popover-list">
        ${domainPresets.map((f) =>
          renderPopoverItem(
            f.string || f.name,
            query.presetFilters.includes(f.name),
            () => callbacks.onTogglePreset(f.name),
          ),
        )}
      </ul>
      <div class="sum-popover-custom">
        <strong class="sum-popover-custom-title">Custom filter</strong>
        <select
          class="sum-popover-select"
          @change=${(e: Event) => {
            const next = (e.target as HTMLSelectElement).value;
            callbacks.onCustomFieldChange(next, filterOperatorsForField(next, filterFields)[0] ?? "=");
          }}
        >
          ${filterFields.map((f) =>
            html`<option value=${f.name} selected=${f.name === field ? "selected" : undefined}>${f.string || f.name}</option>`,
          )}
        </select>
        <select class="sum-popover-select" @change=${(e: Event) => callbacks.onCustomOpChange((e.target as HTMLSelectElement).value)}>
          ${operators.map((op) =>
            html`<option value=${op} selected=${op === customOp ? "selected" : undefined}>${op}</option>`,
          )}
        </select>
        <input
          type="text"
          class="sum-popover-input"
          placeholder="Value"
          value=${customValue}
          @input=${(e: Event) => callbacks.onCustomValueInput(inputValueFromEvent(e))}
        />
        <button type="button" class="sum-btn sum-btn--secondary" @click=${() => callbacks.onApplyCustom()}>Apply</button>
      </div>
    </div>
  `;
}

export interface GroupPopoverState {
  query: CollectionQuery;
  groupPresets: SwcSearchFilter[];
  groupByFields: SwcArchField[];
}

export interface GroupPopoverCallbacks {
  onToggleGroupBy: (field: string) => void;
}

export function renderGroupPopover(state: GroupPopoverState, callbacks: GroupPopoverCallbacks): TemplateResult {
  const { query, groupPresets, groupByFields } = state;

  return html`
    <div class="sum-popover sum-popover--group" @click=${(e: Event) => e.stopPropagation()}>
      <h3 class="sum-popover-heading">Group By</h3>
      <ul class="sum-popover-list">
        ${groupPresets.map((f) =>
          renderPopoverItem(
            f.string || f.name,
            query.groupBy.includes(f.groupBy!),
            () => callbacks.onToggleGroupBy(f.groupBy!),
          ),
        )}
        ${groupByFields.map((f) =>
          renderPopoverItem(
            f.string || f.name,
            query.groupBy.includes(f.name),
            () => callbacks.onToggleGroupBy(f.name),
          ),
        )}
      </ul>
    </div>
  `;
}

export interface FavoriteEntry {
  id: number;
  name: string;
  isShared?: boolean;
  search?: string;
  filter?: string;
  domain?: string;
  groupBy?: string;
}

export interface FavoritesPopoverState {
  favorites: FavoriteEntry[];
  saveName: string;
  saveShared: boolean;
  savingFavorite: boolean;
}

export interface FavoritesPopoverCallbacks {
  onApplyFavorite: (fav: FavoriteEntry) => void;
  onDeleteFavorite: (id: number) => void;
  onSaveNameInput: (name: string) => void;
  onSaveSharedChange: (shared: boolean) => void;
  onSaveFavorite: () => void;
}

export function renderFavoritesPopover(
  state: FavoritesPopoverState,
  callbacks: FavoritesPopoverCallbacks,
): TemplateResult {
  const { favorites, saveName, saveShared, savingFavorite } = state;

  return html`
    <div class="sum-popover sum-popover--favorites" @click=${(e: Event) => e.stopPropagation()}>
      <h3 class="sum-popover-heading">Favorites</h3>
      <ul class="sum-popover-list">
        ${favorites.map(
          (f) => html`<li class="sum-popover-fav">
            <button type="button" class="sum-popover-item" @click=${() => callbacks.onApplyFavorite(f)}>
              <span class="sum-popover-item-label">${f.name}${f.isShared ? " (shared)" : ""}</span>
            </button>
            <button type="button" class="sum-popover-fav-delete" @click=${() => void callbacks.onDeleteFavorite(f.id)} title="Delete">×</button>
          </li>`,
        )}
      </ul>
      <div class="sum-popover-custom">
        <input
          type="text"
          class="sum-popover-input"
          placeholder="Favorite name"
          value=${saveName}
          @input=${(e: Event) => callbacks.onSaveNameInput(inputValueFromEvent(e))}
        />
        <label class="sum-popover-check">
          <input
            type="checkbox"
            ?checked=${saveShared}
            @change=${(e: Event) => callbacks.onSaveSharedChange((e.target as HTMLInputElement).checked)}
          />
          Share with all users
        </label>
        <button
          type="button"
          class="sum-btn sum-btn--secondary"
          disabled=${savingFavorite ? "disabled" : undefined}
          @click=${() => void callbacks.onSaveFavorite()}
        >
          Save current search
        </button>
      </div>
    </div>
  `;
}

export function renderActionsPopover(payload: SwcWorkspacePayload, fieldsCsv: string): TemplateResult {
  const entries = buildReportActionEntries(payload, fieldsCsv);

  return html`
    <div class="sum-popover sum-popover--actions" @click=${(e: Event) => e.stopPropagation()}>
      <h3 class="sum-popover-heading">Actions</h3>
      <ul class="sum-popover-menu">
        ${entries.map((entry) => html`<li class="sum-popover-menu-item">${entry.node}</li>`)}
      </ul>
    </div>
  `;
}

export { presetDomainFilters, presetGroupByFilters };
