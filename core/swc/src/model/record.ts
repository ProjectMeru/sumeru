import type { RpcService } from "../services/rpc.js";
import type { BusService } from "../services/bus.js";
import { RECORD_UPDATED } from "../constants/routes.js";
import { SwcError } from "../runtime/error.js";
import type { SwcArchField } from "../types/workspace.js";

export type RecordUpdatedPayload = { model: string; id?: number; recordId?: number };

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

  /**
   * Removes client-only display fields (e.g. `partner_id_name`) before an RPC
   * write/create. The server whitelists real model fields and rejects unknown
   * keys, so these display helpers must never be sent.
   */
  private serverValues(values: Record<string, unknown>): Record<string, unknown> {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(values)) {
      if (k.endsWith("_name") || k.endsWith("_names")) continue;
      out[k] = v;
    }
    return out;
  }

  async save(rec: SwcRecord): Promise<number> {
    if (rec.id <= 0) {
      const newId = await this.rpc.create(rec.model, this.serverValues(rec.data));
      rec.clearDirty();
      return newId;
    }
    if (!rec.isDirty()) return rec.id;
    await this.rpc.write(rec.model, [rec.id], this.serverValues(rec.dirtyValues()));
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
    return this.rpc.create(rec.model, this.serverValues(values));
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

/** Shared client cache keyed by model:id; use via env.services.record. */
export class RecordService {
  private readonly store: RecordStore;
  private readonly bus: BusService;
  private readonly cache = new Map<string, SwcRecord>();
  // ponytail: FIFO eviction at 64 entries; upgrade to LRU if profiling shows churn.
  private static readonly maxCache = 64;

  constructor(rpc: RpcService, bus: BusService) {
    this.store = new RecordStore(rpc);
    this.bus = bus;
    bus.subscribe(RECORD_UPDATED, (payload) => {
      const msg = payload as RecordUpdatedPayload;
      if (!msg?.model) return;
      const rid = msg.id ?? msg.recordId;
      if (rid != null && rid > 0) {
        this.invalidate(msg.model, rid);
      } else {
        this.invalidate(msg.model);
      }
    });
  }

  private cacheKey(model: string, id: number): string {
    return `${model}:${id}`;
  }

  fromPayload(model: string, id: number, data: Record<string, unknown>): SwcRecord {
    if (id > 0) {
      const hit = this.cache.get(this.cacheKey(model, id));
      if (hit) {
        hit.data = { ...data };
        hit.clearDirty();
        return hit;
      }
    }
    const rec = this.store.fromPayload(model, id, data);
    if (id > 0) {
      this.remember(model, id, rec);
    }
    return rec;
  }

  get(model: string, id: number): SwcRecord | undefined {
    if (id <= 0) return undefined;
    return this.cache.get(this.cacheKey(model, id));
  }

  invalidate(model: string, id?: number): void {
    if (id != null && id > 0) {
      this.cache.delete(this.cacheKey(model, id));
      return;
    }
    for (const key of [...this.cache.keys()]) {
      if (key.startsWith(`${model}:`)) {
        this.cache.delete(key);
      }
    }
  }

  private remember(model: string, id: number, rec: SwcRecord): void {
    while (this.cache.size >= RecordService.maxCache) {
      const first = this.cache.keys().next().value;
      if (first === undefined) break;
      this.cache.delete(first);
    }
    this.cache.set(this.cacheKey(model, id), rec);
  }

  private emitUpdated(model: string, id: number): void {
    this.bus.emit(RECORD_UPDATED, { model, id });
  }

  async save(rec: SwcRecord): Promise<number> {
    const id = await this.store.save(rec);
    if (id > 0) {
      this.remember(rec.model, id, rec);
    }
    this.emitUpdated(rec.model, id > 0 ? id : rec.id);
    return id;
  }

  async unlink(rec: SwcRecord): Promise<void> {
    const id = rec.id;
    await this.store.unlink(rec);
    if (id > 0) {
      this.invalidate(rec.model, id);
    }
    this.emitUpdated(rec.model, id);
  }

  async duplicate(rec: SwcRecord, omit: string[] = ["id"]): Promise<number> {
    const newId = await this.store.duplicate(rec, omit);
    this.emitUpdated(rec.model, newId);
    return newId;
  }

  applyOnchange(rec: SwcRecord, field: string): Promise<OnchangeResult | null> {
    return this.store.applyOnchange(rec, field);
  }

  validate(rec: SwcRecord, requiredFields: string[]): void {
    this.store.validate(rec, requiredFields);
  }
}
