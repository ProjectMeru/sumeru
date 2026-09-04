import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";
import { AsyncFieldController, recordDisplayName } from "./field-async.js";
import { fieldDomain, isFieldReadonly } from "../model/modifiers.js";
import type { FieldWidgetProps } from "./field-props.js";
import { inputValueFromEvent } from "./field-events.js";
import { SelectCreateDialog } from "../views/dialogs/select-create-dialog.js";

export class Many2OneField extends SwcComponent<FieldWidgetProps> {
  private query = "";
  private suggestions: Record<string, unknown>[] = [];
  private open = false;
  private highlightIndex = 0;
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

  private recordId(): number {
    const raw = this.props.record.get(this.props.field.name);
    const id = typeof raw === "number" ? raw : Number(raw);
    return Number.isFinite(id) && id > 0 ? id : 0;
  }

  private closeSuggestions(): void {
    this.open = false;
    this.asyncCtrl.refresh();
  }

  private inputDisplay(recordDisplay: string): string {
    return this.query !== "" || this.open ? this.query : recordDisplay;
  }

  private hasExactNameMatch(): boolean {
    const q = this.query.trim().toLowerCase();
    if (!q) return false;
    return this.suggestions.some((row) => String(row.name ?? "").trim().toLowerCase() === q);
  }

  private async search(query: string): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const comodel = this.comodel();
    if (!comodel) return;
    const baseDomain = fieldDomain(this.props.field, this.props.record) ?? [];
    const domain = query ? [...baseDomain, ["name", "ilike", query]] : baseDomain;
    this.suggestions = await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20);
    this.open = true;
    this.highlightIndex = 0;
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

  private pick(row: Record<string, unknown>): void {
    const { field, record } = this.props;
    record.set(field.name, row.id);
    record.set(`${field.name}_name`, row.name);
    record.notifyFieldChange(field.name);
    this.query = "";
    this.open = false;
    this.suggestions = [];
    this.asyncCtrl.refresh();
  }

  private async createFromQuery(): Promise<void> {
    const name = this.query.trim();
    const comodel = this.comodel();
    if (!name || !comodel) return;
    const id = await this.env.services.rpc.create(comodel, { name });
    if (typeof id === "number" && id > 0) {
      this.pick({ id, name });
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
      onSelect: (row) => this.pick(row),
    });
  }

  private openRelatedRecord(): void {
    const comodel = this.comodel();
    const recordId = this.recordId();
    if (!comodel || recordId <= 0) return;
    void this.env.services.action.applyCallResult({
      open: { model: comodel, recordId, target: "dialog" },
    });
  }

  private onKeydown(event: KeyboardEvent): void {
    if (!this.open) return;
    const createVisible = this.query.trim() !== "" && !this.hasExactNameMatch();
    const optionCount = this.suggestions.length + (createVisible ? 1 : 0);
    if (optionCount === 0) return;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      this.highlightIndex = (this.highlightIndex + 1) % optionCount;
      this.asyncCtrl.refresh();
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      this.highlightIndex = (this.highlightIndex - 1 + optionCount) % optionCount;
      this.asyncCtrl.refresh();
    } else if (event.key === "Enter") {
      event.preventDefault();
      if (this.highlightIndex < this.suggestions.length) {
        const row = this.suggestions[this.highlightIndex];
        if (row) this.pick(row);
      } else if (createVisible) {
        void this.createFromQuery();
      }
    } else if (event.key === "Escape") {
      this.closeSuggestions();
    }
  }

  override template() {
    const { field, record, readonly } = this.props;
    const display = recordDisplayName(record, field.name);
    const id = fieldInputId(field);
    const hasValue = this.recordId() > 0;
    const placeholder = fieldPlaceholder(field);
    const createVisible = this.query.trim() !== "" && !this.hasExactNameMatch();

    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(field, fieldReadonlyValue(display, placeholder), { labelFor: false });
    }

    return renderFieldShell(
      field,
      html`<div class=${this.open ? "sum-m2o-wrap sum-m2o-wrap--open" : "sum-m2o-wrap"}>
        <input
          id=${id}
          class="sum-field-input"
          name=${field.name}
          placeholder=${placeholder}
          value=${this.inputDisplay(display)}
          autocomplete="off"
          @input=${(event: Event) => this.onInput(event)}
          @focus=${() => {
            if (!this.open) void this.search(this.query);
          }}
          @keydown=${(event: Event) => this.onKeydown(event as KeyboardEvent)}
        />
        <div class="sum-m2o-actions">
          ${hasValue
            ? html`<button
                type="button"
                class="sum-m2o-open-btn"
                title="Open record"
                tabindex="-1"
                @click=${() => this.openRelatedRecord()}
              >
                ↗
              </button>`
            : ""}
          <button
            type="button"
            class="sum-m2o-dropdown-btn"
            title="Open suggestions"
            aria-label="Open suggestions"
            aria-expanded=${this.open ? "true" : "false"}
            tabindex="-1"
            @click=${() => this.toggleDropdown()}
          >
            <span class="sum-m2o-caret" aria-hidden="true"></span>
          </button>
        </div>
        ${this.open
          ? html`<ul class="sum-m2o-suggest">
              ${this.suggestions.map(
                (row, index) => html`<li>
                  <button
                    type="button"
                    class=${index === this.highlightIndex
                      ? "sum-m2o-option sum-m2o-option--active"
                      : "sum-m2o-option"}
                    @click=${() => this.pick(row)}
                  >
                    ${String(row.name ?? row.id)}
                  </button>
                </li>`,
              )}
              ${createVisible
                ? html`<li>
                    <button
                      type="button"
                      class=${this.highlightIndex === this.suggestions.length
                        ? "sum-m2o-suggest-create sum-m2o-option--active"
                        : "sum-m2o-suggest-create"}
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
