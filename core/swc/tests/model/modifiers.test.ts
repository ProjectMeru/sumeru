import { describe, it, expect } from "vitest";
import { fieldModifiers, isFieldReadonly, isFieldVisible, fieldDomain, createDefaults, isButtonVisible, evalModifierExpr } from "../../src/model/modifiers.js";
import { SwcRecord } from "../../src/model/record.js";

describe("field modifiers", () => {
  it("respects static invisible flag", () => {
    expect(isFieldVisible({ name: "x", invisible: true })).toBe(false);
    expect(isFieldVisible({ name: "x" })).toBe(true);
  });

  it("merges record overrides", () => {
    const field = { name: "x", readonly: false };
    const mods = fieldModifiers(field, {
      modifierOverrides: new Map([["x", { readonly: true }]]),
    } as never);
    expect(mods.readonly).toBe(true);
  });

  it("evaluates modifier expressions", () => {
    const field = { name: "amount", invisible_expr: "state == 'done'" };
    const visible = fieldModifiers(field, {
      data: { state: "draft" },
      modifierOverrides: new Map(),
    } as never);
    expect(visible.invisible).toBe(false);
    const hidden = fieldModifiers(field, {
      data: { state: "done" },
      modifierOverrides: new Map(),
    } as never);
    expect(hidden.invisible).toBe(true);
  });

  it("isFieldReadonly applies view and expression flags", () => {
    const field = { name: "name", readonly_expr: "state == 'done'" };
    expect(isFieldReadonly(field, undefined, true)).toBe(true);
    expect(
      isFieldReadonly(
        field,
        { data: { state: "draft" }, modifierOverrides: new Map() } as never,
        false,
      ),
    ).toBe(false);
    expect(
      isFieldReadonly(
        field,
        { data: { state: "done" }, modifierOverrides: new Map() } as never,
        false,
      ),
    ).toBe(true);
  });

  it("fieldDomain parses static domain and substitutes placeholders", () => {
    const record = new SwcRecord("m", 1, { partner_id: 42 });
    const fromStatic = fieldDomain({ name: "x", options: { domain: '[["partner_id","=","$partner_id$"]]' } }, record);
    expect(fromStatic).toEqual([["partner_id", "=", 42]]);
    record.fieldDomains.set("x", [["id", ">", 0]]);
    expect(fieldDomain({ name: "x" }, record)).toEqual([["id", ">", 0]]);
  });

  it("createDefaults collects arch defaults", () => {
    expect(createDefaults([{ name: "a", default: 1 }, { name: "b" }])).toEqual({ a: 1 });
  });

  it("isButtonVisible respects invisible_expr", () => {
    const record = { data: { state: "done" }, modifierOverrides: new Map() } as never;
    expect(isButtonVisible({ name: "x", invisible_expr: "state == 'done'" }, record)).toBe(false);
    expect(isButtonVisible({ name: "x", invisible: true }, record)).toBe(false);
  });

  it("evalModifierExpr returns undefined on bad expressions", () => {
    const record = { data: {}, modifierOverrides: new Map() } as never;
    expect(evalModifierExpr("not valid ++", record)).toBeUndefined();
  });
});
