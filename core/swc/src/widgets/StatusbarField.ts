import { SwcComponent } from "../runtime/component.js";
import { html } from "../template/html.js";
import type { SwcArchField } from "../types/workspace.js";
import type { SwcRecord } from "../model/record.js";
import { AsyncFieldController } from "./field-async.js";

interface FieldProps {
  field: SwcArchField;
  record: SwcRecord;
  readonly: boolean;
}

interface StageRow {
  id: number | string;
  label: string;
}

function isClickable(field: SwcArchField): boolean {
  const opt = field.options?.clickable;
  return opt !== "0" && opt !== "false";
}

export class StatusbarField extends SwcComponent<FieldProps> {
  private stages: StageRow[] = [];
  private loaded = false;
  private readonly asyncCtrl = new AsyncFieldController(this);

  setup(): void {
    void this.loadStages();
  }

  onWillUnmount(): void {
    this.asyncCtrl.cancel();
  }

  private async loadStages(): Promise<void> {
    const gen = this.asyncCtrl.begin();
    const { field } = this.props;

    if (field.selection?.length) {
      this.stages = field.selection.map(([value, label]) => ({ id: value, label }));
      this.loaded = true;
      this.asyncCtrl.finish(gen);
      return;
    }

    const comodel = field.relation ?? field.options?.relation ?? "";
    if (!comodel) {
      const fallback = (field.options?.states ?? "draft,done").split(",").map((s) => s.trim()).filter(Boolean);
      this.stages = fallback.map((s) => ({ id: s, label: s }));
      this.loaded = true;
      this.asyncCtrl.finish(gen);
      return;
    }

    const rows = await this.env.services.rpc.searchRead(comodel, [], ["id", "name", "sequence"], 200);
    rows.sort((a, b) => Number(a.sequence ?? 0) - Number(b.sequence ?? 0));
    this.stages = rows.map((row) => ({
      id: Number(row.id),
      label: String(row.name ?? row.id),
    }));
    this.loaded = true;
    this.asyncCtrl.finish(gen);
  }

  private currentId(): string | number {
    const { field, record } = this.props;
    const raw = record.get(field.name);
    if (raw == null || raw === "") return "";
    return field.type === "many2one" || field.relation ? Number(raw) : String(raw);
  }

  template() {
    const { field, record, readonly } = this.props;
    const current = this.currentId();
    const clickable = isClickable(field) && !readonly && !field.readonly;

    return html`<div class="sum-statusbar-stages" role="group" aria-label=${field.string ?? field.name}>
      ${this.stages.map((stage) => {
        const active = stage.id === current || String(stage.id) === String(current);
        const stageClass = active
          ? "sum-statusbar-stage sum-statusbar-stage--current"
          : "sum-statusbar-stage";
        return html`<button type="button" class=${stageClass} disabled=${!clickable ? "disabled" : undefined} @click=${() => {
          if (!clickable) return;
          record.set(field.name, stage.id);
          if (field.type === "many2one" || field.relation) {
            record.set(`${field.name}_name`, stage.label);
          }
          this.asyncCtrl.refresh();
        }}>${stage.label}</button>`;
      })}
      ${!this.loaded ? html`<span class="sum-statusbar-chip">Loading…</span>` : ""}
    </div>`;
  }
}
