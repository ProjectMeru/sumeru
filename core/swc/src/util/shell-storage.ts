/** Persisted shell preferences (legacy-compatible localStorage keys). */

export const KEY_SIDEBAR = "sum.shell.sidebarCollapsed";
export const KEY_ACTIVITY_WIDTH = "sum.shell.activityWidthPx";
export const KEY_ACTIVITY_HIDDEN = "sum.shell.activityHidden";

function storageWarn(op: string, key: string, err: unknown): void {
  console.warn(`shell-storage ${op} failed`, key, err);
}

export function readBool(key: string): boolean {
  try {
    return localStorage.getItem(key) === "1";
  } catch (err) {
    storageWarn("readBool", key, err);
    return false;
  }
}

export function writeBool(key: string, value: boolean): void {
  try {
    localStorage.setItem(key, value ? "1" : "0");
  } catch (err) {
    storageWarn("writeBool", key, err);
  }
}

export function readActivityWidth(): number {
  try {
    const n = parseInt(localStorage.getItem(KEY_ACTIVITY_WIDTH) ?? "", 10);
    if (n >= 200 && n <= 520) return n;
  } catch (err) {
    storageWarn("readActivityWidth", KEY_ACTIVITY_WIDTH, err);
  }
  return 300;
}

export function writeActivityWidth(px: number): void {
  try {
    localStorage.setItem(KEY_ACTIVITY_WIDTH, String(Math.round(px)));
  } catch (err) {
    storageWarn("writeActivityWidth", KEY_ACTIVITY_WIDTH, err);
  }
}

export function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    const value: unknown = JSON.parse(raw);
    if (Array.isArray(fallback)) {
      return (Array.isArray(value) ? value : fallback) as T;
    }
    if (value !== null && typeof value === "object" && !Array.isArray(value)) {
      return value as T;
    }
    return fallback;
  } catch (err) {
    storageWarn("readJSON", key, err);
    return fallback;
  }
}

export function writeJSON(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch (err) {
    storageWarn("writeJSON", key, err);
  }
}
