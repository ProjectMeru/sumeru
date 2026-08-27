import type { RpcService } from "../services/rpc.js";
import { SwcError } from "../runtime/error.js";
import type { SwcArchField } from "../types/workspace.js";

export interface OnchangeResult {
  value?: Record<string, unknown>;
  warning?: { title: string; message: string };
  domain?: Record<string, unknown[]>;
}

export class SwcRecord {
  readonly model: string;
  readonly id: number;
  data: Record<string, unknown>;
  private dirty = new Set<string>();
  /** Client-side field domains from onchange (field name → domain). */
  fieldDomains = new Map<string, unknown[]>();
  /** Dynamic modifier overrides from onchange or eval. */
  modifierOverrides = new Map<string, Partial<Pick<SwcArchField, "invisible" | "readonly" | "required">>>();

  /** Optional callback after a field value changes (onchange RPC). */
  onFieldChange?: (field: string) => void;

  constructor(model: string, id: number, data: Record<string, unknown>) {
    this.model = model;
    this.id = id;
    this.data = { ...data };
  }

  get(field: string): unknown {
    return this.data[field];
  }

  set(field: string, value: unknown): void {
    this.data[field] = value;
    this.dirty.add(field);
  }

  /** Notify listeners that a field finished editing (triggers onchange RPC). */
  notifyFieldChange(field: string): void {
    this.onFieldChange?.(field);
  }

  isDirty(): boolean {
    return this.dirty.size > 0;
  }

  dirtyValues(): Record<string, unknown> {
    const out: Record<string, unknown> = {};
    for (const k of this.dirty) {
      out[k] = this.data[k];
    }
    return out;
  }

  clearDirty(): void {
    this.dirty.clear();
  }

  values(): Record<string, unknown> {
    return { ...this.data };
  }
}

export class RecordStore {
  private readonly rpc: RpcService;

  constructor(rpc: RpcService) {
    this.rpc = rpc;
  }

  fromPayload(model: string, id: number, data: Record<string, unknown>): SwcRecord {
    return new SwcRecord(model, id, data);
  }

  async save(rec: SwcRecord): Promise<number> {
    if (rec.id <= 0) {
      const newId = await this.rpc.create(rec.model, rec.data);
      rec.clearDirty();
      return newId;
    }
    if (!rec.isDirty()) return rec.id;
    await this.rpc.write(rec.model, [rec.id], rec.dirtyValues());
    rec.clearDirty();
    return rec.id;
  }

  async unlink(rec: SwcRecord): Promise<void> {
    if (rec.id <= 0) return;
    await this.rpc.unlink(rec.model, [rec.id]);
  }

  async duplicate(rec: SwcRecord, omit: string[] = ["id"]): Promise<number> {
    const values: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(rec.data)) {
      if (omit.includes(k)) continue;
      values[k] = v;
    }
    return this.rpc.create(rec.model, values);
  }

  async applyOnchange(rec: SwcRecord, field: string): Promise<OnchangeResult | null> {
    try {
      const result = (await this.rpc.onchange(rec.model, rec.values(), field)) as OnchangeResult;
      if (result.value) {
        for (const [k, v] of Object.entries(result.value)) {
          rec.set(k, v);
        }
      }
      if (result.domain) {
        for (const [k, domain] of Object.entries(result.domain)) {
          rec.fieldDomains.set(k, domain as unknown[]);
        }
      }
      return result;
    } catch (err) {
      if (err instanceof SwcError && err.code === "rpc_error") return null;
      throw err;
    }
  }

  validate(rec: SwcRecord, requiredFields: string[]): void {
    for (const f of requiredFields) {
      const v = rec.get(f);
      if (v == null || v === "") {
        throw new SwcError(`Field ${f} is required`, "validation");
      }
    }
  }
}
