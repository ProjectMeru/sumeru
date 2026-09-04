import { afterEach, describe, expect, it, vi } from "vitest";
import { Many2ManyTagsField } from "../../src/widgets/Many2ManyTagsField.js";
import { SwcRecord } from "../../src/model/record.js";
import type { SwcEnv } from "../../src/runtime/env.js";

function env(extra?: Partial<SwcEnv["services"]>): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      rpc: {
        searchRead: vi.fn().mockResolvedValue([
          { id: 1, name: "Red" },
          { id: 2, name: "Blue" },
        ]),
        create: vi.fn().mockResolvedValue(50),
      },
      ...extra,
    },
  } as never;
}

describe("Many2ManyTagsField", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("searches and adds a tag from suggestions", async () => {
    const record = new SwcRecord("m", 1, { tag_ids: [] });
    const comp = new Many2ManyTagsField(
      { field: { name: "tag_ids", relation: "my.tag", string: "Tags" }, record, readonly: false },
      env(),
    );
    comp.callSetup();
    document.body.append(comp.render());
    const input = comp.rootElement!.querySelector(".sum-field-input") as HTMLInputElement;
    input.value = "Re";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-option")).toBeTruthy());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-option")?.click();
    expect(record.get("tag_ids")).toEqual([1]);
    expect(comp.rootElement!.textContent).toContain("Red");
  });

  it("Search more opens SelectCreate and appends", async () => {
    const searchRead = vi.fn().mockResolvedValue([{ id: 3, name: "Green" }]);
    const record = new SwcRecord("m", 1, { tag_ids: [] });
    const comp = new Many2ManyTagsField(
      { field: { name: "tag_ids", relation: "my.tag" }, record, readonly: false },
      env({ rpc: { searchRead, create: vi.fn() } as never }),
    );
    comp.callSetup();
    document.body.append(comp.render());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-dropdown-btn")?.click();
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest-more")).toBeTruthy());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-suggest-more")?.click();
    await vi.waitFor(() => expect(document.querySelector(".sum-select-create-item")).toBeTruthy());
    (document.querySelector(".sum-select-create-item") as HTMLButtonElement).click();
    expect(record.get("tag_ids")).toEqual([3]);
  });

  it("Create from query adds new tag", async () => {
    const create = vi.fn().mockResolvedValue(50);
    const searchRead = vi.fn().mockResolvedValue([]);
    const record = new SwcRecord("m", 1, { tag_ids: [] });
    const comp = new Many2ManyTagsField(
      { field: { name: "tag_ids", relation: "my.tag" }, record, readonly: false },
      env({ rpc: { searchRead, create } as never }),
    );
    comp.callSetup();
    document.body.append(comp.render());
    const input = comp.rootElement!.querySelector(".sum-field-input") as HTMLInputElement;
    input.value = "Purple";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.waitFor(() => expect(comp.rootElement!.querySelector(".sum-m2o-suggest-create")).toBeTruthy());
    comp.rootElement!.querySelector<HTMLButtonElement>(".sum-m2o-suggest-create")?.click();
    await vi.waitFor(() => expect(create).toHaveBeenCalledWith("my.tag", { name: "Purple" }));
    await vi.waitFor(() => expect(record.get("tag_ids")).toEqual([50]));
  });

  it("readonly shows tags without input", () => {
    const record = new SwcRecord("m", 1, {
      tag_ids: [1],
      tag_ids_names: [{ id: 1, name: "Red" }],
    });
    const comp = new Many2ManyTagsField(
      { field: { name: "tag_ids", relation: "my.tag" }, record, readonly: true },
      env(),
    );
    comp.callSetup();
    const el = comp.render();
    expect(el.textContent).toContain("Red");
    expect(el.querySelector(".sum-field-input")).toBeNull();
  });
});
