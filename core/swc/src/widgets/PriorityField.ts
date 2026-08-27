import { SwcComponent } from "../runtime/component.js";
import { html, type TemplateResult } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import {
  fieldInputId,
  fieldLabelId,
  fieldReadonlyValue,
  renderFieldShell,
} from "./field-shell.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

function priorityMode(field: SwcArchField): "stars" | "select" {
  const mode = (field.options?.mode ?? field.options?.display ?? "stars").toLowerCase();
  return mode === "select" || mode === "dropdown" ? "select" : "stars";
}

function selectionOptions(field: SwcArchField): Array<{ value: string; label: string }> {
  if (!field.selection?.length) {
    return [
      { value: "0", label: "Low" },
      { value: "1", label: "Medium" },
      { value: "2", label: "High" },
    ];
  }
  return field.selection.map(([value, label]) => ({ value, label }));
}

function currentValue(field: SwcArchField, record: SwcRecord): string {
  const raw = record.get(field.name);
  if (raw == null || raw === "") return selectionOptions(field)[0]?.value ?? "0";
  return String(raw);
}

function numericLevel(value: string): number {
  const n = Number.parseInt(value, 10);
  return Number.isNaN(n) ? 0 : Math.max(0, n);
}

function starCount(field: SwcArchField): number {
  const fromOpt = Number(field.options?.stars ?? field.options?.max ?? 0);
  if (fromOpt > 0) return Math.min(Math.max(fromOpt, 1), 5);
  const maxLevel = selectionOptions(field).length - 1;
  return Math.max(Math.min(maxLevel, 4), 3);
}

export class PriorityField extends SwcComponent<FieldProps> {
  template() {
    const { field, record, readonly } = this.props;
    const options = selectionOptions(field);
    const value = currentValue(field, record);
    const mode = priorityMode(field);

    if (readonly || field.readonly) {
      const label = options.find((o) => o.value === value)?.label ?? value;
      if (mode === "select") {
        return renderFieldShell(field, fieldReadonlyValue(label), { labelFor: false });
      }
      return renderFieldShell(field, this.renderStars(numericLevel(value), true), { labelFor: false });
    }

    if (mode === "select") {
      const id = fieldInputId(field);
      return renderFieldShell(
        field,
        html`<select
          id=${id}
          class="sum-field-select sum-priority-select"
          name=${field.name}
          @change=${(ev: Event) =>
            record.set(field.name, (ev.target as HTMLSelectElement).value)}
        >
          ${options.map(
            (opt) =>
              html`<option value=${opt.value} selected=${value === opt.value ? "selected" : ""}>
                ${opt.label}
              </option>`,
          )}
        </select>`,
        { labelFor: id },
      );
    }

    return renderFieldShell(
      field,
      this.renderStars(numericLevel(value), false, (level) => {
        record.set(field.name, String(level));
      }),
      { labelFor: false },
    );
  }

  private starButtons(
    level: number,
    disabled: boolean,
    onPick?: (level: number) => void,
  ): TemplateResult[] {
    const { field } = this.props;
    const options = selectionOptions(field);
    const count = starCount(field);
    const capped = Math.min(level, count);
    const out: TemplateResult[] = [];

    for (let i = 0; i < count; i += 1) {
      const starIndex = i + 1;
      const filled = starIndex <= capped;
      const opt = options[Math.min(starIndex, options.length - 1)];
      const click = () => {
        if (disabled) return;
        const next = capped === starIndex ? starIndex - 1 : starIndex;
        onPick?.(Math.max(0, next));
      };
      if (filled) {
        out.push(html`<button type="button" class="sum-priority-star sum-priority-star--on" disabled=${disabled ? "disabled" : undefined} title=${opt?.label ?? `Level ${starIndex}`} aria-label=${opt?.label ?? `Priority ${starIndex}`} @click=${click}>★</button>`);
      } else {
        out.push(html`<button type="button" class="sum-priority-star" disabled=${disabled ? "disabled" : undefined} title=${opt?.label ?? `Level ${starIndex}`} aria-label=${opt?.label ?? `Priority ${starIndex}`} @click=${click}>★</button>`);
      }
    }
    return out;
  }

  private renderStars(
    level: number,
    disabled: boolean,
    onPick?: (level: number) => void,
  ): TemplateResult {
    const { field } = this.props;
    return html`<div class="sum-priority-stars" role="group" aria-labelledby=${fieldLabelId(field)}>
      ${this.starButtons(level, disabled, onPick)}
    </div>`;
  }
}
