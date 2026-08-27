import { describe, it, expect } from "vitest";
import { fieldModifiers, isFieldReadonly, isFieldVisible } from "../../src/model/modifiers.js";

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
});
