import { describe, expect, it, vi, afterEach } from "vitest";
import { Many2OneField } from "../../src/widgets/Many2OneField.js";
import { SwcRecord } from "../../src/model/record.js";
import type { SwcEnv } from "../../src/runtime/env.js";

function env(): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      rpc: {
        searchRead: vi.fn().mockResolvedValue([
          { id: 1, name: "Acme" },
          { id: 2, name: "Beta" },
        ]),
      },
      dialog: { openHost: vi.fn() },
    },
  } as never;
}

describe("Many2OneField", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("searches and picks a suggestion", async () => {
    const record = new SwcRecord("m", 1, { partner_id: null });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner", string: "Partner" }, record, readonly: false },
      env(),
    );
    comp.callSetup();
    document.body.append(comp.render());
    const input = comp.rootElement!.querySelector(".sum-field-input") as HTMLInputElement;
    input.value = "Ac";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest")).toBeTruthy());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-option")?.click();
    comp.patch();
    expect(record.get("partner_id")).toBe(1);
    expect(record.get("partner_id_name")).toBe("Acme");
  });

  it("keyboard navigation selects highlighted row", async () => {
    const record = new SwcRecord("m", 1, { partner_id: null });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner" }, record, readonly: false },
      env(),
    );
    comp.callSetup();
    document.body.append(comp.render());
    const input = comp.rootElement!.querySelector(".sum-field-input") as HTMLInputElement;
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-option")).toBeTruthy());
    input.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
    expect(record.get("partner_id")).toBe(1);
  });

  it("readonly shows display name", () => {
    const record = new SwcRecord("m", 1, { partner_id: 5, partner_id_name: "Vendor" });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner", string: "Partner" }, record, readonly: true },
      env(),
    );
    comp.callSetup();
    const el = comp.render();
    expect(el.textContent).toContain("Vendor");
  });
});
