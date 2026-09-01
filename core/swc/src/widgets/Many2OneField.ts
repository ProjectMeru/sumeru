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
  private suggestions: Record<string, unknown>[] = [];
  private open = false;
  private highlightIndex = 0;
  private readonly asyncCtrl = new AsyncFieldController(this);

  override onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async search(query: string): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const comodel = this.props.field.relation ?? this.props.field.options?.relation ?? "";
    if (!comodel) return;
    const baseDomain = fieldDomain(this.props.field, this.props.record) ?? [];
    const domain = query ? [...baseDomain, ["name", "ilike", query]] : baseDomain;
    this.suggestions = await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20);
    this.open = true;
    this.highlightIndex = 0;
    this.asyncCtrl.commitIfCurrent(gen);
  }

  private pick(row: Record<string, unknown>): void {
    const { field, record } = this.props;
    record.set(field.name, row.id);
    record.set(`${field.name}_name`, row.name);
    record.notifyFieldChange(field.name);
    this.open = false;
    this.asyncCtrl.refresh();
  }

  private openSelectCreate(): void {
    const { field, record } = this.props;
    const comodel = field.relation ?? field.options?.relation ?? "";
    if (!comodel) return;
    SelectCreateDialog.open(this.env, {
      comodel,
      title: field.string ?? field.name,
      domain: fieldDomain(field, record) ?? [],
      onSelect: (row) => this.pick(row),
    });
  }

  private onKeydown(event: KeyboardEvent): void {
    if (!this.open || this.suggestions.length === 0) return;
    if (event.key === "ArrowDown") {
      event.preventDefault();
      this.highlightIndex = (this.highlightIndex + 1) % this.suggestions.length;
      this.asyncCtrl.refresh();
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      this.highlightIndex =
        (this.highlightIndex - 1 + this.suggestions.length) % this.suggestions.length;
      this.asyncCtrl.refresh();
    } else if (event.key === "Enter") {
      event.preventDefault();
      const row = this.suggestions[this.highlightIndex];
      if (row) this.pick(row);
    } else if (event.key === "Escape") {
      this.open = false;
      this.asyncCtrl.refresh();
    }
  }

  override template() {
    const { field, record, readonly } = this.props;
    const display = recordDisplayName(record, field.name);
    const id = fieldInputId(field);

    const placeholder = fieldPlaceholder(field);

    if (isFieldReadonly(field, record, readonly)) {
      return renderFieldShell(field, fieldReadonlyValue(display, placeholder), { labelFor: false });
    }

    return renderFieldShell(
      field,
      html`<div class="sum-m2o-wrap">
        <input
          id=${id}
          class="sum-field-input"
          name=${field.name}
          placeholder=${placeholder}
          value=${display}
          autocomplete="off"
          @input=${(event: Event) => void this.search(inputValueFromEvent(event))}
          @keydown=${(event: Event) => this.onKeydown(event as KeyboardEvent)}
        />
        <button type="button" class="sum-m2o-search-btn" title="Search records" @click=${() => this.openSelectCreate()}>…</button>
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
            </ul>`
          : ""}
      </div>`,
      { labelFor: id },
    );
  }
}
