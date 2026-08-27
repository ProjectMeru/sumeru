import { SwcComponent } from "../runtime/component.js";
import { html, type TemplateResult } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { fieldControl, renderFieldShell } from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

interface O2MLine {
  id: number;
  data: Record<string, unknown>;
}

let tempLineId = -1;

function nextTempId(): number {
  tempLineId -= 1;
  return tempLineId;
}

function inverseFieldName(parentModel: string): string {
  const part = parentModel.split(".").pop() ?? "parent";
  return `${part}_id`;
}

function columnsForField(field: SwcArchField): SwcArchField[] {
  return field.subview?.fields ?? [];
}

function columnNames(cols: SwcArchField[]): string[] {
  return cols.map((c) => c.name);
}

function parseCellValue(col: SwcArchField, raw: string): unknown {
  if (raw === "") return null;
  if (col.type === "integer") return Number.parseInt(raw, 10);
  if (col.type === "float" || col.type === "numeric") return Number.parseFloat(raw);
  if (col.type === "boolean") return raw === "true" || raw === "1";
  return raw;
}

function displayCellValue(col: SwcArchField, line: Record<string, unknown>): string {
  const raw = line[col.name];
  if (raw == null) return "";
  const named = line[`${col.name}_name`];
  if (named != null && String(named) !== "") return String(named);
  if (col.type === "boolean") {
    return raw === true || raw === 1 || raw === "1" || raw === "true" ? "Yes" : "No";
  }
  return String(raw);
}

export class One2ManyField extends SwcComponent<FieldProps> {
  private lines: O2MLine[] = [];
  private loaded = false;
  private saving = false;
  private readonly asyncCtrl = new AsyncFieldController(this);
  private readonly writeTimers = new Map<string, ReturnType<typeof setTimeout>>();

  setup(): void {
    void this.loadLines();
  }

  onWillUnmount(): void {
    this.asyncCtrl.cancel();
    for (const t of this.writeTimers.values()) clearTimeout(t);
    this.writeTimers.clear();
  }

  private comodel(): string {
    const { field } = this.props;
    return field.relation ?? field.options?.relation ?? "";
  }

  private inverse(): string {
    const { field, record } = this.props;
    return field.options?.inverse ?? inverseFieldName(record.model);
  }

  private editable(): boolean {
    const { field, record, readonly } = this.props;
    if (readonly || field.readonly) return false;
    if (record.id <= 0) return false;
    const mode = field.subview?.editable ?? "bottom";
    return mode === "bottom" || mode === "top";
  }

  private async loadLines(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field, record } = this.props;
    const comodel = this.comodel();
    const cols = columnsForField(field);
    if (!comodel || record.id <= 0 || cols.length === 0) {
      this.loaded = true;
      this.asyncCtrl.finish(gen);
      return;
    }
    const inv = this.inverse();
    const names = ["id", ...columnNames(cols)];
    const rows = await this.env.services.rpc.searchRead(
      comodel,
      [[inv, "=", record.id]],
      names,
      200,
    );
    this.lines = (rows ?? []).map((row) => ({
      id: Number(row.id ?? 0),
      data: { ...row },
    }));
    this.loaded = true;
    this.asyncCtrl.finish(gen);
  }

  private lineById(id: number): O2MLine | undefined {
    return this.lines.find((l) => l.id === id);
  }

  private scheduleWrite(lineId: number, col: SwcArchField, value: unknown): void {
    const key = `${lineId}:${col.name}`;
    const prev = this.writeTimers.get(key);
    if (prev) clearTimeout(prev);
    this.writeTimers.set(
      key,
      setTimeout(() => {
        this.writeTimers.delete(key);
        void this.persistCell(lineId, col, value);
      }, 350),
    );
  }

  private async persistCell(lineId: number, col: SwcArchField, value: unknown): Promise<void> {
    if (lineId <= 0) return;
    const comodel = this.comodel();
    if (!comodel) return;
    this.saving = true;
    this.asyncCtrl.refresh();
    try {
      await this.env.services.rpc.write(comodel, [lineId], { [col.name]: value });
      const line = this.lineById(lineId);
      if (line) line.data[col.name] = value;
    } finally {
      this.saving = false;
      this.asyncCtrl.refresh();
    }
  }

  private async createLine(lineId: number, col: SwcArchField, value: unknown): Promise<void> {
    const { record } = this.props;
    const comodel = this.comodel();
    const line = this.lineById(lineId);
    if (!comodel || !line || line.id > 0) return;

    line.data[col.name] = value;
    this.saving = true;
    this.asyncCtrl.refresh();
    try {
      const vals: Record<string, unknown> = { ...line.data, [this.inverse()]: record.id };
      delete vals.id;
      const newId = await this.env.services.rpc.create(comodel, vals);
      line.id = newId;
      line.data.id = newId;
    } finally {
      this.saving = false;
      this.asyncCtrl.refresh();
    }
  }

  private onCellInput(lineId: number, col: SwcArchField, raw: unknown): void {
    const value =
      typeof raw === "boolean" ? raw : parseCellValue(col, String(raw ?? ""));
    const line = this.lineById(lineId);
    if (!line) return;
    line.data[col.name] = value;
    if (line.id <= 0) {
      void this.createLine(lineId, col, value);
      return;
    }
    this.scheduleWrite(line.id, col, value);
  }

  private async addRowViaDialog(): Promise<void> {
    const cols = columnsForField(this.props.field);
    if (cols.length === 0) return;
    const dialog = this.env.services.dialog;
    if (dialog) {
      const ok = await dialog.confirm(
        "Add line",
        `Add a new line to ${this.props.field.string ?? this.props.field.name}?`,
      );
      if (!ok) return;
    }
    this.addRow();
  }

  private addRow(): void {
    const id = nextTempId();
    this.lines = [...this.lines, { id, data: {} }];
    this.asyncCtrl.refresh();
  }

  private async deleteRow(lineId: number): Promise<void> {
    const comodel = this.comodel();
    const line = this.lineById(lineId);
    if (!line) return;
    if (line.id > 0 && comodel) {
      this.saving = true;
      this.asyncCtrl.refresh();
      try {
        await this.env.services.rpc.unlink(comodel, [line.id]);
      } finally {
        this.saving = false;
      }
    }
    this.lines = this.lines.filter((l) => l.id !== lineId);
    this.asyncCtrl.refresh();
  }

  private renderCellEditor(col: SwcArchField, line: O2MLine): ReturnType<typeof html> {
    const val = String(line.data[col.name] ?? "");
    const readonly = !this.editable();

    if (readonly) {
      return html`<span>${displayCellValue(col, line.data)}</span>`;
    }

    if (col.type === "boolean") {
      const checked = line.data[col.name] === true || line.data[col.name] === 1;
      return fieldControl(
        html`<input
          type="checkbox"
          class="sum-field-input"
          checked=${checked ? "checked" : ""}
          @change=${(ev: Event) =>
            this.onCellInput(line.id, col, (ev.target as HTMLInputElement).checked)}
        />`,
        true,
      );
    }

    if (col.selection?.length) {
      return fieldControl(
        html`<select
          class="sum-field-select"
          @change=${(ev: Event) =>
            this.onCellInput(line.id, col, (ev.target as HTMLSelectElement).value)}
        >
          <option value="">—</option>
          ${col.selection.map(
            ([v, label]) =>
              html`<option value=${v} selected=${val === v ? "selected" : ""}>${label}</option>`,
          )}
        </select>`,
        true,
      );
    }

    const inputType =
      col.type === "integer" || col.type === "float" || col.type === "numeric"
        ? "number"
        : col.type === "date"
          ? "date"
          : "text";

    return fieldControl(
      html`<input
        type=${inputType}
        class="sum-field-input"
        value=${val}
        @input=${(ev: Event) =>
          this.onCellInput(line.id, col, (ev.target as HTMLInputElement).value)}
      />`,
      true,
    );
  }

  private renderLineRow(line: O2MLine, cols: SwcArchField[], canEdit: boolean): TemplateResult {
    const cells: TemplateResult[] = cols.map(
      (col) => html`<td>${this.renderCellEditor(col, line)}</td>`,
    );
    if (canEdit) {
      cells.push(html`<td class="sum-o2m-col-actions"><button type="button" .sum-o2m-delete-btn data-line-id=${String(line.id)} title="Remove line">×</button></td>`);
    }
    return html`<tr class="sum-o2m-row">${cells}</tr>`;
  }

  private onTableClick(ev: Event): void {
    const btn = (ev.target as HTMLElement).closest(".sum-o2m-delete-btn");
    if (!btn) return;
    const id = Number(btn.getAttribute("data-line-id"));
    if (!Number.isFinite(id)) return;
    void this.deleteRow(id);
  }

  template() {
    const { field, record } = this.props;
    const label = field.string ?? field.name;
    const cols = columnsForField(field);
    const canEdit = this.editable();
    const emptyMsg = !this.loaded
      ? "Loading…"
      : record.id <= 0
        ? "Save the record before adding lines."
        : cols.length === 0
          ? "No columns configured."
          : "No lines";

    return renderFieldShell(
      field,
      html`<div class="sum-o2m-table-wrap">
        <div class="sum-o2m-title">${label}${this.saving ? " (saving…)" : ""}</div>
        <table class="sum-o2m-table">
          <thead>
            <tr>
              ${cols.map((col) => html`<th>${col.string ?? col.name}</th>`)}
              ${canEdit ? html`<th class="sum-o2m-col-actions"></th>` : ""}
            </tr>
          </thead>
          <tbody @click=${(ev: Event) => this.onTableClick(ev)}>
            ${this.lines.length === 0
              ? html`<tr>
                  <td colspan=${String(cols.length + (canEdit ? 1 : 0))}>${emptyMsg}</td>
                </tr>`
              : this.lines.map((line) => this.renderLineRow(line, cols, canEdit))}
          </tbody>
        </table>
        ${canEdit && cols.length > 0
          ? html`<button type="button" class="sum-o2m-add-row" @click=${() => void this.addRowViaDialog()}>
              + Add a line
            </button>`
          : ""}
        ${!canEdit && record.id <= 0 && !this.props.readonly
          ? html`<p class="sum-o2m-hint">Save the parent record before editing lines.</p>`
          : ""}
      </div>`,
      { layout: "stack", showLabel: false },
    );
  }
}
