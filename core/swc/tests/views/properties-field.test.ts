import { describe, expect, it } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import { PropertiesField } from "../../src/widgets/PropertiesField.js";
import type { SwcArchField } from "../../src/types/workspace.js";
import type { SwcEnv } from "../../src/runtime/env.js";

function env(): SwcEnv {
  return { bootstrap: {} as never, services: {} as never } as SwcEnv;
}

describe("PropertiesField", () => {
  it("renders one input per definition and not a JSON editor", () => {
    const field: SwcArchField = {
      name: "props",
      string: "Properties",
      type: "properties",
      widget: "properties",
      options: {
        schema: JSON.stringify([
          { key: "color", label: "Color", value_type: "char" },
          { key: "qty", label: "Qty", value_type: "integer" },
        ]),
      },
    };
    const record = new SwcRecord("demo.model", 1, { props: { color: "red", qty: 2 } });
    const comp = new PropertiesField({ field, record, readonly: false }, env());
    comp.callSetup();
    const el = comp.render();
    expect(el.textContent).toContain("Color");
    expect(el.textContent).toContain("Qty");
    expect(el.querySelector("textarea")).toBeNull();
    expect(el.querySelectorAll("input").length).toBe(2);
  });

  it("shows an empty state when no definitions exist", () => {
    const field: SwcArchField = {
      name: "props",
      string: "Properties",
      type: "properties",
      widget: "properties",
      options: { schema: "[]" },
    };
    const record = new SwcRecord("demo.model", 1, { props: {} });
    const comp = new PropertiesField({ field, record, readonly: false }, env());
    comp.callSetup();
    expect(comp.render().textContent).toContain("No properties defined");
  });
});
