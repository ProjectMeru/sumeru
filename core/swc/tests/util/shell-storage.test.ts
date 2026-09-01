import { describe, expect, it, beforeEach } from "vitest";
import {
  KEY_ACTIVITY_WIDTH,
  KEY_SIDEBAR,
  readActivityWidth,
  readBool,
  readJSON,
  writeActivityWidth,
  writeBool,
  writeJSON,
} from "../../src/util/shell-storage.js";

describe("shell-storage", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("readBool and writeBool round-trip", () => {
    expect(readBool(KEY_SIDEBAR)).toBe(false);
    writeBool(KEY_SIDEBAR, true);
    expect(readBool(KEY_SIDEBAR)).toBe(true);
    writeBool(KEY_SIDEBAR, false);
    expect(readBool(KEY_SIDEBAR)).toBe(false);
  });

  it("readActivityWidth clamps and defaults", () => {
    expect(readActivityWidth()).toBe(300);
    writeActivityWidth(999);
    expect(readActivityWidth()).toBe(300);
    writeActivityWidth(240);
    expect(readActivityWidth()).toBe(240);
  });

  it("readJSON and writeJSON round-trip arrays", () => {
    expect(readJSON<string[]>("apps", [])).toEqual([]);
    writeJSON("apps", ["a", "b"]);
    expect(readJSON<string[]>("apps", [])).toEqual(["a", "b"]);
    writeJSON("apps", { not: "array" });
    expect(readJSON<string[]>("apps", ["fallback"])).toEqual(["fallback"]);
  });
});
