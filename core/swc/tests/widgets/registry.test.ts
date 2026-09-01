import { describe, expect, it, vi } from "vitest";
import { registerDefaultWidgets, resolveFieldWidget, renderField, instantiateFieldWidget } from "../../src/widgets/registry.js";
import { SwcRecord } from "../../src/model/record.js";
import type { SwcEnv } from "../../src/runtime/env.js";

describe("field widget registry", () => {
  it("resolveFieldWidget maps types and aliases", () => {
    expect(resolveFieldWidget({ name: "x", type: "boolean" })).toBe("boolean");
    expect(resolveFieldWidget({ name: "x", type: "many2many" })).toBe("many2many_tags");
    expect(resolveFieldWidget({ name: "x", widget: "progressbar" })).toBe("progress");
    expect(resolveFieldWidget({ name: "x", widget: "phone" })).toBe("phone");
  });

  it("instantiateFieldWidget and renderField mount widgets", () => {
    registerDefaultWidgets();
    const env = { bootstrap: {} as never, services: { rpc: { searchRead: vi.fn() } } } as unknown as SwcEnv;
    const record = new SwcRecord("m", 1, { name: "A" });
    const widget = instantiateFieldWidget(env, { name: "name", type: "char" }, record, false);
    expect(widget.render().classList.contains("sum-field-widget")).toBe(true);
    widget.destroy();
    expect(renderField(env, { name: "name", type: "char" }, record, false).querySelector(".sum-field-input")).toBeTruthy();
  });

  it("falls back to default widget for unknown widget keys", () => {
    registerDefaultWidgets();
    const env = { bootstrap: {} as never, services: {} } as unknown as SwcEnv;
    const record = new SwcRecord("m", 1, { ref: "x" });
    const el = renderField(env, { name: "ref", widget: "unknown_widget_key" }, record, false);
    expect(el.querySelector(".sum-field-input")).toBeTruthy();
  });
});
