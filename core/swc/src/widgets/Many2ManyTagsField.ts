import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import { fieldInputId, fieldPlaceholder, renderFieldShell } from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";
import type { FieldWidgetProps } from "./field-props.js";
import type { SwcRecord } from "../model/record.js";
import { inputValueFromEvent } from "./field-events.js";
import { fieldDomain, isFieldReadonly } from "../model/modifiers.js";
import { SelectCreateDialog } from "../views/dialogs/select-create-dialog.js";

interface TagRow {
  id: number;
  name: string;
}

function tagIds(record: SwcRecord, fieldName: string): number[] {
  const rawValue = record.get(fieldName);
  if (!Array.isArray(rawValue)) return [];
  return rawValue.map((v) => Number(v)).filter((n) => !Number.isNaN(n));
}

function tagNamesFromRecord(record: SwcRecord, fieldName: string): Map<number, string> {
  const out = new Map<number, string>();
  const rawValue = record.get(`${fieldName}_names`);
  if (Array.isArray(rawValue)) {
    for (const item of rawValue) {
      if (item && typeof item === "object") {
        const row = item as Record<string, unknown>;
        const id = Number(row.id);
        if (!Number.isNaN(id)) out.set(id, String(row.name ?? id));
      }
    }
  }
  return out;
}

export class Many2ManyTagsField extends SwcComponent<FieldWidgetProps> {
  private query = "";
  private suggestions: TagRow[] = [];
  private open = false;
  private nameCache = new Map<number, string>();
  private readonly asyncCtrl = new AsyncFieldController(this);

  private readonly onDocClick = (event: MouseEvent): void => {
    if (!this.open) return;
    const target = event.target;
    if (!(target instanceof Node)) return;
    if (this.rootElement?.contains(target)) return;
    this.closeSuggestions();
  };

  private readonly onDocKeydown = (event: KeyboardEvent): void => {
    if (event.key === "Escape" && this.open) {
      event.preventDefault();
      this.closeSuggestions();
    }
  };

  override onMount(): void {
    document.addEventListener("click", this.onDocClick, true);
    document.addEventListener("keydown", this.onDocKeydown);
  }

  override onWillUnmount(): void {
    document.removeEventListener("click", this.onDocClick, true);
    document.removeEventListener("keydown", this.onDocKeydown);
    this.asyncCtrl.cancel();
  }

  private comodel(): string {
    return this.props.field.relation ?? this.props.field.options?.relation ?? "";
  }

  private closeSuggestions(): void {
    this.open = false;
    this.asyncCtrl.refresh();
  }

  private selectedTags(): TagRow[] {
    const { field, record } = this.props;
    const ids = tagIds(record, field.name);
    const names = tagNamesFromRecord(record, field.name);
    return ids.map((id) => {
      const fromCache = this.nameCache.get(id);
      if (fromCache) return { id, name: fromCache };
      const fromRecord = names.get(id);
      if (fromRecord) return { id, name: fromRecord };
      return { id, name: `#${id}` };
    });
  }

  private setIds(ids: number[], names?: Map<number, string>): void {
    const { field, record } = this.props;
    record.set(field.name, ids);
    if (names) {
      for (const [id, name] of names) this.nameCache.set(id, name);
      record.set(
        `${field.name}_names`,
        ids.map((id) => ({ id, name: this.nameCache.get(id) ?? `#${id}` })),
      );
    }
    record.notifyFieldChange(field.name);
    this.asyncCtrl.refresh();
  }

  private addTag(row: TagRow): void {
    const selected = this.selectedTags();
    if (selected.some((t) => t.id === row.id)) {
      this.query = "";
      this.closeSuggestions();
      return;
    }
    this.nameCache.set(row.id, row.name);
    this.setIds(
      [...selected.map((t) => t.id), row.id],
      new Map([[row.id, row.name]]),
    );
    this.query = "";
    this.suggestions = [];
    this.open = false;
    this.asyncCtrl.refresh();
  }

  private hasExactNameMatch(): boolean {
    const q = this.query.trim().toLowerCase();
    if (!q) return false;
    return this.suggestions.some((row) => row.name.trim().toLowerCase() === q);
  }

  private async search(query: string): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const comodel = this.comodel();
    if (!comodel) return;
    const selected = new Set(this.selectedTags().map((t) => t.id));
    const baseDomain = fieldDomain(this.props.field, this.props.record) ?? [];
    const domain = query ? [...baseDomain, ["name", "ilike", query]] : baseDomain;
    const rows =
      (await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20)) ?? [];
    this.suggestions = rows
      .map((row) => ({ id: Number(row.id), name: String(row.name ?? row.id) }))
      .filter((row) => !Number.isNaN(row.id) && !selected.has(row.id));
    this.open = true;
    this.asyncCtrl.commitIfCurrent(gen);
  }

  private onInput(event: Event): void {
    this.query = inputValueFromEvent(event);
    void this.search(this.query);
  }

  private toggleDropdown(): void {
    if (this.open) {
      this.closeSuggestions();
      return;
    }
    void this.search(this.query);
  }

  private async createFromQuery(): Promise<void> {
    const name = this.query.trim();
    const comodel = this.comodel();
    if (!name || !comodel) return;
    const id = await this.env.services.rpc.create(comodel, { name });
    if (typeof id === "number" && id > 0) {
      this.addTag({ id, name });
    }
  }

  private openSelectCreate(): void {
    const { field, record } = this.props;
    const comodel = this.comodel();
    if (!comodel) return;
    const initialQuery = this.query;
    this.open = false;
    this.asyncCtrl.refresh();
    SelectCreateDialog.open(this.env, {
      comodel,
      title: field.string ?? field.name,
      domain: fieldDomain(field, record) ?? [],
      initialQuery,
      onSelect: (row) => {
        this.addTag({ id: Number(row.id), name: String(row.name ?? row.id) });
      },
    });
  }

  override template() {
    const { field, record, readonly } = this.props;
    const selected = this.selectedTags();
    const createVisible = this.query.trim() !== "" && !this.hasExactNameMatch();

    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(
        field,
        html`<div class="sum-multi-select-tags sum-multi-select-tags--readonly sum-field-tags">
          ${selected.map(
            (tag) =>
              html`<span class="sum-multi-select-tag"
                ><span class="sum-multi-select-tag-label">${tag.name}</span></span
              >`,
          )}
        </div>`,
        { labelFor: false },
      );
    }

    const id = fieldInputId(field);
    const placeholder = fieldPlaceholder(field) || "Search…";

    return renderFieldShell(
      field,
      html`<div class="sum-m2m-wrap sum-multi-select-box">
        <div class="sum-multi-select-tags sum-field-tags">
          ${selected.map(
            (tag) =>
              html`<span class="sum-multi-select-tag">
                <span class="sum-multi-select-tag-label">${tag.name}</span>
                <button
                  type="button"
                  class="sum-multi-select-remove"
                  aria-label="Remove"
                  @click=${() =>
                    this.setIds(selected.filter((t) => t.id !== tag.id).map((t) => t.id))}
                >
                  ×
                </button>
              </span>`,
          )}
        </div>
        <div class="sum-m2m-input-row">
          <input
            id=${id}
            class="sum-field-input"
            placeholder=${placeholder}
            value=${this.query}
            autocomplete="off"
            @input=${(event: Event) => this.onInput(event)}
            @focus=${() => {
              if (!this.open) void this.search(this.query);
            }}
          />
          <button
            type="button"
            class="sum-m2o-dropdown-btn"
            title="Open suggestions"
            aria-label="Open suggestions"
            tabindex="-1"
            @click=${() => this.toggleDropdown()}
          >
            <span class="sum-m2o-caret" aria-hidden="true"></span>
          </button>
        </div>
        ${this.open
          ? html`<ul class="sum-m2o-suggest">
              ${this.suggestions.map(
                (row) => html`<li>
                  <button type="button" class="sum-m2o-option" @click=${() => this.addTag(row)}>
                    ${row.name}
                  </button>
                </li>`,
              )}
              ${createVisible
                ? html`<li>
                    <button
                      type="button"
                      class="sum-m2o-suggest-create"
                      @click=${() => void this.createFromQuery()}
                    >
                      Create "${this.query.trim()}"
                    </button>
                  </li>`
                : ""}
              <li>
                <button
                  type="button"
                  class="sum-m2o-suggest-more"
                  @click=${() => this.openSelectCreate()}
                >
                  Search more…
                </button>
              </li>
            </ul>`
          : ""}
      </div>`,
      { labelFor: id },
    );
  }
}
