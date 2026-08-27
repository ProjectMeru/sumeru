import { describe, expect, it, vi } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import { StatusbarField } from "../../src/widgets/StatusbarField.js";
import type { SwcArchField } from "../../src/types/workspace.js";
import type { SwcEnv } from "../../src/runtime/env.js";

describe("StatusbarField", () => {
  it("renders selection states from field.selection", () => {
    const env = {
      bootstrap: {} as never,
      services: { rpc: { searchRead: vi.fn() } },
    } as unknown as SwcEnv;
    const record = new SwcRecord("my.module", 1, { state: "draft" });
    const field: SwcArchField = {
      name: "state",
      widget: "statusbar",
      type: "selection",
      selection: [
        ["draft", "Draft"],
        ["done", "Done"],
      ],
    };
    const comp = new StatusbarField({ field, record, readonly: false }, env);
    comp.setup?.();
    const el = comp.render();
    expect(el.textContent).toContain("Draft");
    expect(el.textContent).toContain("Done");
    expect(el.innerHTML).toContain("sum-statusbar-stage");
  });

  it("loads M2O stages via RPC", async () => {
    const searchRead = vi.fn().mockResolvedValue([
      { id: 1, name: "New", sequence: 1 },
      { id: 2, name: "Won", sequence: 2 },
    ]);
    const env = {
      bootstrap: {} as never,
      services: { rpc: { searchRead } },
    } as unknown as SwcEnv;
    const record = new SwcRecord("crm.lead", 1, { stage_id: 2 });
    const field: SwcArchField = {
      name: "stage_id",
      widget: "statusbar",
      type: "many2one",
      relation: "crm.stage",
    };
    const comp = new StatusbarField({ field, record, readonly: false }, env);
    comp.setup?.();
    const host = document.createElement("div");
    host.appendChild(comp.render());
    document.body.appendChild(host);
    await new Promise((r) => setTimeout(r, 20));
    comp.patch();
    expect(searchRead).toHaveBeenCalledWith("crm.stage", [], ["id", "name", "sequence"], 200);
    expect(comp.el?.textContent).toContain("New");
    expect(comp.el?.textContent).toContain("Won");
    host.remove();
  });
});
