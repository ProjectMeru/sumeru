import type { SwcRecord } from "./record.js";

/**
 * Child records (one2many lines) staged on an unsaved parent record.
 *
 * One2ManyField keeps its in-memory rows here so FormView can create them
 * after the parent record is saved (the server ignores one2many fields on
 * create, so lines must be inserted separately with the parent id).
 */
export interface PendingChildRecord {
  fieldName: string;
  comodel: string;
  inverse: string;
  values: Record<string, unknown>;
}

const store = new WeakMap<SwcRecord, Map<string, PendingChildRecord[]>>();

function byField(record: SwcRecord): Map<string, PendingChildRecord[]> {
  let map = store.get(record);
  if (!map) {
    map = new Map();
    store.set(record, map);
  }
  return map;
}

export function getPendingChildren(
  record: SwcRecord,
  fieldName: string,
): PendingChildRecord[] | undefined {
  return store.get(record)?.get(fieldName);
}

export function setPendingChildren(
  record: SwcRecord,
  fieldName: string,
  children: PendingChildRecord[],
): void {
  const map = byField(record);
  if (children.length === 0) {
    map.delete(fieldName);
  } else {
    map.set(fieldName, children);
  }
}

export function takePendingChildren(record: SwcRecord): PendingChildRecord[] {
  const map = store.get(record);
  if (!map) return [];
  const out: PendingChildRecord[] = [];
  for (const children of map.values()) {
    out.push(...children);
  }
  store.delete(record);
  return out;
}
