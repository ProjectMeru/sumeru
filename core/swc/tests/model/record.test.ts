import { describe, it, expect, vi } from "vitest";
import { RecordStore, SwcRecord } from "../../src/model/record.ts";
import { SwcError } from "../../src/runtime/error.js";

describe("SwcRecord", () => {
  it("tracks dirty fields", () => {
    const rec = new SwcRecord("test.model", 1, { name: "A" });
    rec.set("name", "B");
    expect(rec.isDirty()).toBe(true);
    expect(rec.dirtyValues()).toEqual({ name: "B" });
    rec.clearDirty();
    expect(rec.isDirty()).toBe(false);
  });

  it("notifyFieldChange invokes callback", () => {
    const rec = new SwcRecord("m", 1, {});
    const fn = vi.fn();
    rec.onFieldChange = fn;
    rec.notifyFieldChange("name");
    expect(fn).toHaveBeenCalledWith("name");
  });

  it("values returns a shallow copy", () => {
    const rec = new SwcRecord("m", 1, { a: 1 });
    const vals = rec.values();
    vals.a = 2;
    expect(rec.get("a")).toBe(1);
  });
});

describe("RecordStore", () => {
  function rpc(overrides: Partial<Record<string, unknown>> = {}) {
    return {
      create: vi.fn().mockResolvedValue(99),
      write: vi.fn().mockResolvedValue(undefined),
      unlink: vi.fn().mockResolvedValue(undefined),
      onchange: vi.fn().mockResolvedValue({ value: { qty: 2 }, domain: { partner_id: [] } }),
      ...overrides,
    };
  }

  it("save creates new records and strips display fields", async () => {
    const mock = rpc();
    const store = new RecordStore(mock as never);
    const rec = store.fromPayload("m", 0, { name: "A", partner_id_name: "X" });
    const id = await store.save(rec);
    expect(id).toBe(99);
    expect(mock.create).toHaveBeenCalledWith("m", { name: "A" });
    expect(rec.isDirty()).toBe(false);
  });

  it("save writes dirty fields for existing records", async () => {
    const mock = rpc();
    const store = new RecordStore(mock as never);
    const rec = store.fromPayload("m", 5, { name: "A" });
    rec.set("name", "B");
    await store.save(rec);
    expect(mock.write).toHaveBeenCalledWith("m", [5], { name: "B" });
  });

  it("save skips write when not dirty", async () => {
    const mock = rpc();
    const store = new RecordStore(mock as never);
    const rec = store.fromPayload("m", 5, { name: "A" });
    await store.save(rec);
    expect(mock.write).not.toHaveBeenCalled();
  });

  it("unlink skips new records", async () => {
    const mock = rpc();
    const store = new RecordStore(mock as never);
    await store.unlink(store.fromPayload("m", 0, {}));
    expect(mock.unlink).not.toHaveBeenCalled();
  });

  it("duplicate omits id by default", async () => {
    const mock = rpc();
    const store = new RecordStore(mock as never);
    const rec = store.fromPayload("m", 1, { id: 1, name: "A" });
    await store.duplicate(rec);
    expect(mock.create).toHaveBeenCalledWith("m", { name: "A" });
  });

  it("applyOnchange merges value and domain", async () => {
    const mock = rpc();
    const store = new RecordStore(mock as never);
    const rec = store.fromPayload("m", 1, { qty: 1 });
    const result = await store.applyOnchange(rec, "qty");
    expect(result?.value).toEqual({ qty: 2 });
    expect(rec.get("qty")).toBe(2);
    expect(rec.fieldDomains.get("partner_id")).toEqual([]);
  });

  it("applyOnchange swallows rpc_error", async () => {
    const mock = rpc({
      onchange: vi.fn().mockRejectedValue(new SwcError("fail", "rpc_error")),
    });
    const store = new RecordStore(mock as never);
    const rec = store.fromPayload("m", 1, {});
    await expect(store.applyOnchange(rec, "x")).resolves.toBeNull();
  });

  it("throws on missing required field", () => {
    const store = new RecordStore({} as never);
    const rec = store.fromPayload("m", 0, {});
    expect(() => store.validate(rec, ["name"])).toThrow();
  });
});
