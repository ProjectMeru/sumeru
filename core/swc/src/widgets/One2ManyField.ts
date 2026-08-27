import { SwcComponent } from "../runtime/component.js";
import { html, type TemplateResult } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import {
  getPendingChildren,
  setPendingChildren,
  type PendingChildRecord,
} from "../model/pending-children.js";
import { fieldControl, renderFieldShell } from "./field-shell.js";
import { AsyncFieldController } from "./field-async.js";
import type { FieldWidgetProps } from "./field-props.js";
import { booleanFromUnknown } from "./field-value.js";
import { checkboxCheckedFromEvent, inputValueFromEvent } from "./field-events.js";
import { isFieldReadonly } from "../model/modifiers.js";

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
  if (col.type === "float" || col.type === "float64" || col.type === "numeric") {
    return Number.parseFloat(raw);
  }
  if (col.type === "boolean") return raw === "true" || raw === "1";
  return raw;
}

function isNumericType(col: SwcArchField): boolean {
  return (
    col.type === "integer" ||
    col.type === "float" ||
    col.type === "float64" ||
    col.type === "numeric"
  );
}

/** Formats a number with thousand separators (e.g. 12000 → "12,000"). */
function formatNumericValue(raw: unknown): string {
  const num = Number(raw);
  if (!Number.isFinite(num)) return raw == null ? "" : String(raw);
  const [intPart, decPart] = String(num).split(".");
  const withSep = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  return decPart !== undefined ? `${withSep}.${decPart}` : withSep;
}

function displayCellValue(col: SwcArchField, line: Record<string, unknown>): string {
  const raw = line[col.name];
  if (raw == null) return "";
  const named = line[`${col.name}_name`];
  if (named != null && String(named) !== "") return String(named);
  if (col.type === "boolean") {
    return booleanFromUnknown(raw) ? "Yes" : "No";
  }
  if (isNumericType(col)) return formatNumericValue(raw);
  return String(raw);
}

/** Server-safe line values: drop id, `*_name`/`*_names` display fields, and empty values. */
function serverLineValues(data: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(data)) {
    if (k === "id") continue;
    if (k.endsWith("_name") || k.endsWith("_names")) continue;
    if (v == null || v === "") continue;
    out[k] = v;
  }
  return out;
}

export class One2ManyField extends SwcComponent<FieldWidgetProps> {
  private lines: O2MLine[] = [];
  private loaded = false;
  private saving = false;
  private readonly asyncCtrl = new AsyncFieldController(this);
  private readonly writeTimers = new Map<string, ReturnType<typeof setTimeout>>();
  private readonly m2oQueries = new Map<string, string>();
  private readonly m2oSuggestions = new Map<string, Record<string, unknown>[]>();
  private m2oOpenKey: string | null = null;
  private m2oPopover: HTMLElement | null = null;
  private m2oSearchSeq = 0;
  private onchangeSeq = 0;

  override setup(): void {
    void this.loadLines();
    document.addEventListener("mousedown", this.onDocumentM2oDown);
  }

  override onWillUnmount(): void {
    this.asyncCtrl.cancel();
    this.m2oSearchSeq += 1;
    document.removeEventListener("mousedown", this.onDocumentM2oDown);
    this.closeM2oPopover();
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
    if (isFieldReadonly(field, record, readonly)) return false;
    const mode = field.subview?.editable ?? "bottom";
    return mode === "bottom" || mode === "top";
  }

  private async loadLines(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field, record } = this.props;
    const comodel = this.comodel();
    const cols = columnsForField(field);
    if (!comodel || cols.length === 0) {
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
      return;
    }
    if (record.id <= 0) {
      this.lines = (getPendingChildren(record, field.name) ?? []).map((child) => ({
        id: nextTempId(),
        data: { ...child.values },
      }));
      this.loaded = true;
      this.asyncCtrl.commitIfCurrent(gen);
      return;
    }
    const inv = this.inverse();
    const names = ["id", ...columnNames(cols)];
    const rows = (await this.env.services.rpc.searchRead(
      comodel,
      [[inv, "=", record.id]],
      names,
      200,
    )) ?? [];
    await this.resolveM2oNames(cols, rows);
    this.lines = rows.map((row) => ({
      id: Number(row.id ?? 0),
      data: { ...row },
    }));
    this.loaded = true;
    this.asyncCtrl.commitIfCurrent(gen);
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
      const vals: Record<string, unknown> = {
        ...serverLineValues(line.data),
        [this.inverse()]: record.id,
      };
      const newId = await this.env.services.rpc.create(comodel, vals);
      line.id = newId;
      line.data.id = newId;
    } finally {
      this.saving = false;
      this.asyncCtrl.refresh();
    }
  }

  private async onCellInput(lineId: number, col: SwcArchField, raw: unknown): Promise<void> {
    const value =
      typeof raw === "boolean" ? raw : parseCellValue(col, String(raw ?? ""));
    const line = this.lineById(lineId);
    if (!line) return;
    line.data[col.name] = value;
    const derived = await this.applyLineOnchange(lineId, col.name);
    if (this.props.record.id <= 0) {
      this.syncPendingChildren();
      return;
    }
    if (line.id <= 0) {
      void this.createLine(lineId, col, value);
      return;
    }
    this.scheduleWrite(line.id, col, value);
    for (const name of derived) {
      const dcol = this.columnByName(name);
      if (dcol) this.scheduleWrite(line.id, dcol, line.data[name]);
    }
  }

  /**
   * Updates a readonly cell in place (without re-rendering the table) so the
   * focused input keeps its caret position after a server onchange.
   */
  private updateReadonlyCell(lineId: number, colName: string): void {
    if (!this.rootElement) return;
    const row = this.rootElement.querySelector<HTMLElement>(`tr[data-line-id="${lineId}"]`);
    if (!row) return;
    const cell = row.querySelector<HTMLElement>(
      `td[data-col="${CSS.escape(colName)}"]`,
    );
    if (!cell) return;
    const line = this.lineById(lineId);
    const col = this.columnByName(colName);
    if (line && col) {
      cell.textContent = displayCellValue(col, line.data);
    }
  }

  private columnByName(name: string): SwcArchField | undefined {
    return columnsForField(this.props.field).find((c) => c.name === name);
  }

  /**
   * Asks the server to recompute derived fields for a line (business rules
   * live server-side). Returns the names of the fields the server changed.
   */
  private async applyLineOnchange(lineId: number, field: string): Promise<string[]> {
    const line = this.lineById(lineId);
    const comodel = this.comodel();
    if (!line || !comodel) return [];
    const seq = ++this.onchangeSeq;
    let result: { value?: Record<string, unknown> } | null = null;
    try {
      result = (await this.env.services.rpc.onchange(
        comodel,
        serverLineValues(line.data),
        field,
      )) as { value?: Record<string, unknown> };
    } catch {
      return [];
    }
    if (seq !== this.onchangeSeq) return [];
    const value = result?.value;
    if (!value) return [];
    const changed: string[] = [];
    for (const [name, v] of Object.entries(value)) {
      if (line.data[name] === v) continue;
      line.data[name] = v;
      changed.push(name);
      const col = this.columnByName(name);
      if (col && col.readonly) this.updateReadonlyCell(lineId, name);
    }
    return changed;
  }

  private pendingChildren(): PendingChildRecord[] {
    const children: PendingChildRecord[] = [];
    const comodel = this.comodel();
    const inverse = this.inverse();
    for (const line of this.lines) {
      if (line.id > 0) continue;
      const values = serverLineValues(line.data);
      if (Object.keys(values).length === 0) continue;
      children.push({ fieldName: this.props.field.name, comodel, inverse, values });
    }
    return children;
  }

  private syncPendingChildren(): void {
    const { record, field } = this.props;
    if (record.id > 0) {
      setPendingChildren(record, field.name, []);
      return;
    }
    setPendingChildren(record, field.name, this.pendingChildren());
  }

  private addRow(): void {
    const id = nextTempId();
    this.lines = [...this.lines, { id, data: {} }];
    this.syncPendingChildren();
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
    this.syncPendingChildren();
    this.asyncCtrl.refresh();
  }

  override patch(): void {
    const root = this.rootElement;
    const active = document.activeElement;
    const wasInside = !!(root && active instanceof HTMLElement && root.contains(active));
    let focusKey: string | null = null;
    let caret: number | null = null;
    if (wasInside && active instanceof HTMLInputElement) {
      focusKey = active.getAttribute("data-cell-key") ?? null;
      caret = active.selectionStart;
    }
    super.patch();
    if (focusKey) {
      const next = this.rootElement?.querySelector<HTMLInputElement>(
        `input[data-cell-key="${CSS.escape(focusKey)}"]`,
      );
      if (next) {
        next.focus();
        if (caret !== null) {
          try {
            next.setSelectionRange(caret, caret);
          } catch {
            // ignore — input type may not support selection
          }
        }
      }
    }
    this.scheduleM2oPopover();
  }

  private cellKey(lineId: number, col: SwcArchField): string {
    return `${lineId}:${col.name}`;
  }

  private m2oKeyParts(key: string): { lineId: number; col: SwcArchField | undefined } {
    const idx = key.indexOf(":");
    const lineId = Number(key.slice(0, idx));
    const colName = key.slice(idx + 1);
    return {
      lineId,
      col: columnsForField(this.props.field).find((c) => c.name === colName),
    };
  }

  private m2oComodel(col: SwcArchField): string {
    return col.relation ?? col.options?.relation ?? "";
  }

  private onM2oInput(lineId: number, col: SwcArchField, raw: string): void {
    const key = this.cellKey(lineId, col);
    this.m2oQueries.set(key, raw);
    if (raw.trim() === "") {
      this.closeM2oPopover();
      return;
    }
    void this.searchM2o(key, lineId, col, raw);
  }

  private onM2oFocus(lineId: number, col: SwcArchField): void {
    const key = this.cellKey(lineId, col);
    if ((this.m2oSuggestions.get(key) ?? []).length > 0) {
      this.m2oOpenKey = key;
      this.scheduleM2oPopover();
    } else {
      this.m2oOpenKey = null;
    }
  }

  private onDocumentM2oDown = (ev: MouseEvent): void => {
    const target = ev.target as HTMLElement | null;
    if (!target) return;
    if (
      this.m2oPopover &&
      !this.m2oPopover.contains(target) &&
      !target.closest("input[data-cell-key]")
    ) {
      this.m2oOpenKey = null;
      this.closeM2oPopover();
    }
  };

  private async searchM2o(
    key: string,
    _lineId: number,
    col: SwcArchField,
    q: string,
  ): Promise<void> {
    const comodel = this.m2oComodel(col);
    if (!comodel) return;
    const seq = ++this.m2oSearchSeq;
    const base = q.trim();
    const domain: unknown[] = base ? [["name", "ilike", `%${base}%`]] : [];
    let rows: Record<string, unknown>[] = [];
    try {
      rows =
        (await this.env.services.rpc.searchRead(comodel, domain, ["id", "name"], 20)) ?? [];
    } catch {
      return;
    }
    if (seq !== this.m2oSearchSeq) return;
    this.m2oSuggestions.set(key, rows);
    this.m2oOpenKey = key;
    this.asyncCtrl.refresh();
  }

  private selectM2o(key: string, row: Record<string, unknown>): void {
    const { lineId, col } = this.m2oKeyParts(key);
    const line = this.lineById(lineId);
    if (!line || !col) return;
    line.data[col.name] = row.id;
    line.data[`${col.name}_name`] = row.name;

    let descCol: SwcArchField | undefined;
    let updatedDescription = false;
    if (col.name === "product_id") {
      descCol = this.columnByName("name");
      if (descCol) {
        line.data["name"] = String(row.name ?? "");
        updatedDescription = true;
      }
    }

    this.m2oQueries.set(key, "");
    this.m2oSuggestions.delete(key);
    this.m2oOpenKey = null;
    this.closeM2oPopover();

    if (this.props.record.id <= 0) {
      this.syncPendingChildren();
    } else if (line.id <= 0) {
      void this.createLine(lineId, col, row.id);
    } else {
      this.scheduleWrite(line.id, col, row.id);
      if (updatedDescription && descCol) {
        this.scheduleWrite(line.id, descCol, line.data["name"]);
      }
    }

    this.asyncCtrl.refresh();
  }

  private scheduleM2oPopover(): void {
    const key = this.m2oOpenKey;
    if (!key) {
      this.closeM2oPopover();
      return;
    }
    const rows = this.m2oSuggestions.get(key) ?? [];
    if (rows.length === 0) {
      this.closeM2oPopover();
      return;
    }
    const anchor = this.rootElement?.querySelector<HTMLElement>(
      `input[data-cell-key="${CSS.escape(key)}"]`,
    );
    if (!anchor) {
      this.closeM2oPopover();
      return;
    }
    this.openM2oPopover(anchor, key, rows);
  }

  private openM2oPopover(anchor: HTMLElement, key: string, rows: Record<string, unknown>[]): void {
    let pop = this.m2oPopover;
    if (!pop) {
      pop = document.createElement("ul");
      pop.className = "sum-m2o-suggest sum-o2m-m2o-popover";
      pop.style.position = "fixed";
      pop.style.zIndex = "1000";
      document.body.appendChild(pop);
      this.m2oPopover = pop;
    }
    pop.textContent = "";
    for (const row of rows) {
      const li = document.createElement("li");
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "sum-m2o-option";
      btn.textContent = String(row.name ?? row.id);
      btn.addEventListener("mousedown", (ev) => {
        ev.preventDefault();
        this.selectM2o(key, row);
      });
      li.appendChild(btn);
      pop.appendChild(li);
    }
    const rect = anchor.getBoundingClientRect();
    const width = Math.max(Math.round(rect.width), 1);
    pop.style.top = `${Math.round(rect.bottom + 2)}px`;
    pop.style.left = `${Math.round(rect.left)}px`;
    pop.style.width = `${width}px`;
    pop.style.minWidth = `${width}px`;
    pop.style.maxWidth = `${width}px`;
  }

  private closeM2oPopover(): void {
    this.m2oPopover?.remove();
    this.m2oPopover = null;
  }

  private async resolveM2oNames(
    cols: SwcArchField[],
    rows: Record<string, unknown>[],
  ): Promise<void> {
    for (const col of cols) {
      if (col.type !== "many2one") continue;
      const comodel = this.m2oComodel(col);
      if (!comodel) continue;
      const ids = rows
        .map((r) => Number(r[col.name]))
        .filter((id) => Number.isFinite(id) && id > 0);
      if (ids.length === 0) continue;
      const uniq = Array.from(new Set(ids));
      let refs: Record<string, unknown>[] = [];
      try {
        refs =
          (await this.env.services.rpc.searchRead(
            comodel,
            [["id", "in", uniq]],
            ["id", "name"],
            uniq.length + 1,
          )) ?? [];
      } catch {
        continue;
      }
      const nameById = new Map<number, string>();
      for (const ref of refs) {
        nameById.set(Number(ref.id), String(ref.name ?? ""));
      }
      for (const r of rows) {
        const id = Number(r[col.name]);
        const name = nameById.get(id);
        if (name !== undefined) r[`${col.name}_name`] = name;
      }
    }
  }

  private renderCellEditor(col: SwcArchField, line: O2MLine): ReturnType<typeof html> {
    const fieldValue = String(line.data[col.name] ?? "");
    const readonly = !this.editable() || col.readonly === true;

    if (readonly) {
      return html`<span>${displayCellValue(col, line.data)}</span>`;
    }

    if (col.type === "many2one") {
      const key = this.cellKey(line.id, col);
      const query = this.m2oQueries.get(key) ?? "";
      const display = displayCellValue(col, line.data);
      const value = query !== "" ? query : display;
      return fieldControl(
        html`<div class="sum-o2m-m2o">
          <input
            type="text"
            class="sum-field-input"
            data-cell-key=${key}
            value=${value}
            autocomplete="off"
            @input=${(event: Event) =>
              this.onM2oInput(line.id, col, inputValueFromEvent(event))}
            @focus=${() => this.onM2oFocus(line.id, col)}
          />
        </div>`,
        true,
      );
    }

    if (col.type === "boolean") {
      const checked = booleanFromUnknown(line.data[col.name]);
      return fieldControl(
        html`<input
          type="checkbox"
          class="sum-field-input"
          checked=${checked ? "checked" : ""}
          @change=${(event: Event) =>
            this.onCellInput(line.id, col, checkboxCheckedFromEvent(event))}
        />`,
        true,
      );
    }

    if (col.selection?.length) {
      return fieldControl(
        html`<select
          class="sum-field-select"
          @change=${(event: Event) =>
            this.onCellInput(line.id, col, inputValueFromEvent(event))}
        >
          <option value="">—</option>
          ${col.selection.map(
            ([v, label]) =>
              html`<option value=${v} selected=${fieldValue === v ? "selected" : ""}>${label}</option>`,
          )}
        </select>`,
        true,
      );
    }

    const isNumeric =
      col.type === "integer" ||
      col.type === "float" ||
      col.type === "float64" ||
      col.type === "numeric";
    const inputType = col.type === "date" ? "date" : "text";
    const inputMode = isNumeric
      ? col.type === "integer"
        ? "numeric"
        : "decimal"
      : "";

    return fieldControl(
      html`<input
        type=${inputType}
        class="sum-field-input"
        data-cell-key=${this.cellKey(line.id, col)}
        value=${fieldValue}
        ${inputMode ? html`inputmode=${inputMode}` : ""}
        @input=${(event: Event) =>
          this.onCellInput(line.id, col, inputValueFromEvent(event))}
      />`,
      true,
    );
  }

  private renderLineRow(line: O2MLine, cols: SwcArchField[], canEdit: boolean): TemplateResult {
    const cells: TemplateResult[] = cols.map(
      (col) => html`<td data-col=${col.name}>${this.renderCellEditor(col, line)}</td>`,
    );
    if (canEdit) {
      cells.push(
        html`<td class="sum-o2m-col-actions"><button type="button" class="sum-o2m-delete-btn" data-line-id=${String(line.id)} title="Remove line">×</button></td>`,
      );
    }
    return html`<tr class="sum-o2m-row" data-line-id=${String(line.id)}>${cells}</tr>`;
  }

  private onTableClick(event: Event): void {
    const deleteButton = (event.target as HTMLElement).closest(".sum-o2m-delete-btn");
    if (!deleteButton) return;
    const id = Number(deleteButton.getAttribute("data-line-id"));
    if (!Number.isFinite(id)) return;
    void this.deleteRow(id);
  }

  override template() {
    const { field } = this.props;
    const label = field.string ?? field.name;
    const cols = columnsForField(field);
    const canEdit = this.editable();
    const emptyMsg = !this.loaded
      ? "Loading…"
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
          <tbody @click=${(event: Event) => this.onTableClick(event)}>
            ${this.lines.length === 0
              ? html`<tr>
                  <td colspan=${String(cols.length + (canEdit ? 1 : 0))}>${emptyMsg}</td>
                </tr>`
              : this.lines.map((line) => this.renderLineRow(line, cols, canEdit))}
          </tbody>
        </table>
        ${canEdit && cols.length > 0
          ? html`<button type="button" class="sum-o2m-add-row" @click=${() => this.addRow()}>
              + Add a line
            </button>`
          : ""}
      </div>`,
      { layout: "stack", showLabel: false },
    );
  }
}
