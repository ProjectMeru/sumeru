import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import {
  fieldInputId,
  fieldPlaceholder,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";
import { fieldDomain } from "../model/modifiers.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

export class Many2OneField extends SwcComponent<FieldProps> {
  private suggestions: Record<string, unknown>[] = [];
  private open = false;
  private query = "";
  private readonly asyncCtrl = new AsyncFieldController(this);

  onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  /**
   * The default patch() swaps the whole subtree with replaceChild, which
   * destroys the focused <input>. Restore focus and caret afterwards so the
   * user can keep typing across asynchronous suggestion re-renders.
   */
  override patch(): void {
    const root = this.el;
    const activeEl = document.activeElement;
    const wasFocused = !!(root && activeEl && root.contains(activeEl));
    const input = activeEl instanceof HTMLInputElement ? activeEl : null;
    const caret = input ? input.selectionStart : null;

    super.patch();

    if (wasFocused && caret !== null) {
      const next = this.el?.querySelector<HTMLInputElement>("input");
      if (next) {
        next.focus();
        try {
          next.setSelectionRange(caret, caret);
        } catch {
          // ignore — input type may not support selection
        }
      }
    }
  }

  private async search(q: string): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const comodel = this.props.field.relation ?? this.props.field.options?.relation ?? "";
    if (!comodel) return;
    const baseDomain = fieldDomain(this.props.field, this.props.record) ?? [];
    const domain = q ? [...baseDomain, ["name", "ilike", `%${q}%`]] : baseDomain;
    this.suggestions = await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20);
    this.open = true;
    this.asyncCtrl.finish(gen);
  }

  private onInput(ev: Event): void {
    const input = ev.target as HTMLInputElement;
    this.query = input.value;
    void this.search(input.value);
  }

  private select(row: Record<string, unknown>): void {
    const { field, record } = this.props;
    record.set(field.name, row.id);
    record.set(`${field.name}_name`, row.name);
    record.notifyFieldChange(field.name);
    this.open = false;
    this.query = "";
    this.asyncCtrl.refresh();
  }

  template() {
    const { field, record, readonly } = this.props;
    const display = record.get(`${field.name}_name`) ?? (record.get(field.name) ? `#${record.get(field.name)}` : "");
    const id = fieldInputId(field);

    const placeholder = fieldPlaceholder(field);

    if (readonly || field.readonly) {
      return renderFieldShell(field, fieldReadonlyValue(String(display), placeholder), { labelFor: false });
    }

    // While the user is typing, show the in-progress query; otherwise show the
    // stored record value. The query must survive re-renders or every patch
    // wipes the typed text.
    const value = this.query !== "" ? this.query : String(display);

    return renderFieldShell(
      field,
      html`<div class="sum-m2o-wrap">
        <input
          id=${id}
          class="sum-field-input"
          name=${field.name}
          placeholder=${placeholder}
          value=${value}
          autocomplete="off"
          @input=${(ev: Event) => this.onInput(ev)}
        />
        ${this.open
          ? html`<ul class="sum-m2o-suggest">
              ${this.suggestions.map(
                (row) => html`<li>
                  <button
                    type="button"
                    class="sum-m2o-option"
                    @click=${() => this.select(row)}
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
