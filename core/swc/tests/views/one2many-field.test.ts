import { describe, expect, it, vi, afterEach } from "vitest";
import { One2ManyField } from "../../src/widgets/One2ManyField.js";
import { SwcRecord } from "../../src/model/record.js";
import type { SwcArchField } from "../../src/types/workspace.js";
import type { SwcEnv } from "../../src/runtime/env.js";

function makeEnv(rpc: Partial<SwcEnv["services"]["rpc"]>): SwcEnv {
  return {
    bootstrap: {} as never,
    services: {
      rpc: {
        searchRead: vi.fn().mockResolvedValue([]),
        write: vi.fn().mockResolvedValue(true),
        create: vi.fn().mockResolvedValue(99),
        unlink: vi.fn().mockResolvedValue(true),
        ...rpc,
      },
    } as never,
  };
}

function lineField(): SwcArchField {
  return {
    name: "line_ids",
    string: "Lines",
    type: "one2many",
    relation: "my.module.line",
    options: { inverse: "module_id", relation: "my.module.line" },
    subview: {
      editable: "bottom",
      fields: [
        { name: "name", string: "Description", type: "char" },
        { name: "quantity", string: "Qty", type: "integer" },
      ],
    },
  };
}

async function mountField(
  comp: One2ManyField,
): Promise<{ host: HTMLDivElement; el: HTMLElement }> {
  const host = document.createElement("div");
  host.appendChild(comp.render());
  document.body.appendChild(host);
  await new Promise((r) => setTimeout(r, 20));
  comp.patch();
  return { host, el: comp.el! };
}

describe("One2ManyField", () => {
  afterEach(() => {
    document.body.innerHTML = "";
    vi.useRealTimers();
  });

  it("loads and displays lines", async () => {
    const searchRead = vi.fn().mockResolvedValue([{ id: 10, name: "Line A", quantity: 2 }]);
    const env = makeEnv({ searchRead });
    const record = new SwcRecord("my.module", 1, { name: "Parent" });
    const comp = new One2ManyField({ field: lineField(), record, readonly: true }, env);
    comp.setup();
    const { host } = await mountField(comp);
    expect(searchRead).toHaveBeenCalled();
    expect(comp.el?.textContent).toContain("Line A");
    host.remove();
  });

  it("writes cell changes after debounce", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const searchRead = vi.fn().mockResolvedValue([{ id: 10, name: "Line A", quantity: 1 }]);
    const write = vi.fn().mockResolvedValue(true);
    const env = makeEnv({ searchRead, write });
    const record = new SwcRecord("my.module", 1, { name: "Parent" });
    const comp = new One2ManyField({ field: lineField(), record, readonly: false }, env);
    comp.setup();
    const { host } = await mountField(comp);
    const input = comp.el!.querySelector("tbody input.sum-field-input") as HTMLInputElement;
    expect(input).toBeTruthy();
    input.value = "Updated";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.advanceTimersByTimeAsync(400);
    expect(write).toHaveBeenCalledWith("my.module.line", [10], { name: "Updated" });
    host.remove();
    vi.useRealTimers();
  });

  it("creates a line when editing a new row", async () => {
    const searchRead = vi.fn().mockResolvedValue([]);
    const create = vi.fn().mockResolvedValue(99);
    const env = makeEnv({ searchRead, create });
    const record = new SwcRecord("my.module", 1, { name: "Parent" });
    const comp = new One2ManyField({ field: lineField(), record, readonly: false }, env);
    comp.setup();
    const { host } = await mountField(comp);
    comp.el!.querySelector<HTMLButtonElement>(".sum-o2m-add-row")?.click();
    comp.patch();
    const input = comp.el!.querySelector("tbody input.sum-field-input") as HTMLInputElement;
    input.value = "New line";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await Promise.resolve();
    expect(create).toHaveBeenCalledWith(
      "my.module.line",
      expect.objectContaining({ name: "New line", module_id: 1 }),
    );
    host.remove();
  });

  it("unlinks a line on delete", async () => {
    const searchRead = vi.fn().mockResolvedValue([{ id: 10, name: "Line A", quantity: 1 }]);
    const unlink = vi.fn().mockResolvedValue(true);
    const env = makeEnv({ searchRead, unlink });
    const record = new SwcRecord("my.module", 1, { name: "Parent" });
    const comp = new One2ManyField({ field: lineField(), record, readonly: false }, env);
    comp.setup();
    const { host } = await mountField(comp);
    const btn = comp.el!.querySelector<HTMLButtonElement>(".sum-o2m-delete-btn");
    expect(btn).toBeTruthy();
    expect(btn!.getAttribute("data-line-id")).toBe("10");
    btn!.click();
    await vi.waitFor(() => {
      expect(unlink).toHaveBeenCalledWith("my.module.line", [10]);
    });
    host.remove();
  });
});
