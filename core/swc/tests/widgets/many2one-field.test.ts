import { describe, expect, it, vi, afterEach } from "vitest";
import { Many2OneField } from "../../src/widgets/Many2OneField.js";
import { SwcRecord } from "../../src/model/record.js";
import type { SwcEnv } from "../../src/runtime/env.js";

function env(extra?: Partial<SwcEnv["services"]>): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      rpc: {
        searchRead: vi.fn().mockResolvedValue([
          { id: 1, name: "Acme" },
          { id: 2, name: "Beta" },
        ]),
        create: vi.fn().mockResolvedValue(99),
      },
      dialog: { openHost: vi.fn() },
      action: { applyCallResult: vi.fn().mockResolvedValue(true) },
      ...extra,
    },
  } as never;
}

describe("Many2OneField", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("keeps typed query after async search patch", async () => {
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
    const after = comp.rootElement!.querySelector(".sum-field-input") as HTMLInputElement;
    expect(after.value).toBe("Ac");
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
    expect(el.querySelector(".sum-m2o-dropdown-btn")).toBeNull();
    expect(el.querySelector(".sum-m2o-search-btn")).toBeNull();
  });

  it("chevron opens suggestions", async () => {
    const record = new SwcRecord("m", 1, { partner_id: null });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner", string: "Partner" }, record, readonly: false },
      env(),
    );
    comp.callSetup();
    document.body.append(comp.render());
    expect(comp.rootElement!.querySelector(".sum-m2o-search-btn")).toBeNull();
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-dropdown-btn")?.click();
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest")).toBeTruthy());
  });

  it("Search more footer opens SelectCreate dialog", async () => {
    const record = new SwcRecord("m", 1, { partner_id: null });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner", string: "Partner" }, record, readonly: false },
      env(),
    );
    comp.callSetup();
    document.body.append(comp.render());
    const input = comp.rootElement!.querySelector(".sum-field-input") as HTMLInputElement;
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest-more")).toBeTruthy());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-suggest-more")?.click();
    await vi.waitFor(() => expect(document.querySelector(".sum-select-create")).toBeTruthy());
  });

  it("Create from query calls rpc.create and picks", async () => {
    const create = vi.fn().mockResolvedValue(99);
    const searchRead = vi.fn().mockResolvedValue([{ id: 1, name: "Acme" }]);
    const record = new SwcRecord("m", 1, { partner_id: null });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner", string: "Partner" }, record, readonly: false },
      env({ rpc: { searchRead, create } as never }),
    );
    comp.callSetup();
    document.body.append(comp.render());
    const input = comp.rootElement!.querySelector(".sum-field-input") as HTMLInputElement;
    input.value = "NewCo";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest-create")).toBeTruthy());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-suggest-create")?.click();
    await vi.waitFor(() => expect(create).toHaveBeenCalledWith("res.partner", { name: "NewCo" }));
    await vi.waitFor(() => expect(record.get("partner_id")).toBe(99));
    expect(record.get("partner_id_name")).toBe("NewCo");
  });

  it("open button opens related record when valued", () => {
    const applyCallResult = vi.fn().mockResolvedValue(true);
    const record = new SwcRecord("m", 1, { partner_id: 5, partner_id_name: "Vendor" });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner", string: "Partner" }, record, readonly: false },
      env({ action: { applyCallResult } as never }),
    );
    comp.callSetup();
    document.body.append(comp.render());
    expect(comp.rootElement!.querySelector(".sum-m2o-open-btn")).toBeTruthy();
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-open-btn")?.click();
    expect(applyCallResult).toHaveBeenCalledWith({
      open: { model: "res.partner", recordId: 5, target: "dialog" },
    });
  });

  it("outside click closes suggestions via widget state", async () => {
    const record = new SwcRecord("m", 1, { partner_id: null });
    const comp = new Many2OneField(
      { field: { name: "partner_id", relation: "res.partner" }, record, readonly: false },
      env(),
    );
    comp.callSetup();
    document.body.append(comp.render());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-dropdown-btn")?.click();
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest")).toBeTruthy());
    document.body.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest")).toBeNull());
  });
});
