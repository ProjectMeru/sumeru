import { describe, expect, it, vi } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import { FieldHost } from "../../src/widgets/field-host.js";
import { registerDefaultWidgets } from "../../src/widgets/registry.js";
import type { SwcEnv } from "../../src/runtime/env.js";

registerDefaultWidgets();

describe("FieldHost", () => {
  it("reuses widget instance across renders with same mode", () => {
    const searchRead = vi.fn();
    const env = {
      bootstrap: {} as never,
      services: { rpc: { searchRead } },
    } as unknown as SwcEnv;
    const host = new FieldHost(env);
    const record = new SwcRecord("core.partner", 1, { email: "a@x.com" });
    const field = { name: "email", type: "char" as const };

    const el1 = host.render(field, record, true);
    const el2 = host.render(field, record, true);
    expect(el1).toBe(el2);
    expect((el1.querySelector("input") as HTMLInputElement | null)?.value).toBe("a@x.com");
    expect((el2.querySelector("input") as HTMLInputElement | null)?.value).toBe("a@x.com");
    host.clear();
  });

  it("recreates widget when readonly mode changes", async () => {
    const searchRead = vi.fn().mockResolvedValue([]);
    const env = {
      bootstrap: {} as never,
      services: { rpc: { searchRead } },
    } as unknown as SwcEnv;
    const host = new FieldHost(env);
    const record = new SwcRecord("core.partner", 1, { country_id: 1, country_id_name: "India" });
    const field = { name: "country_id", widget: "selection", relation: "core.country" };

    host.render(field, record, true);
    expect(searchRead).not.toHaveBeenCalled();

    host.render(field, record, false);
    await Promise.resolve();
    expect(searchRead).toHaveBeenCalled();
    host.clear();
  });

  it("invalidates a single field without clearing others", () => {
    const env = {
      bootstrap: {} as never,
      services: { rpc: { searchRead: vi.fn() } },
    } as unknown as SwcEnv;
    const host = new FieldHost(env);
    const record = new SwcRecord("core.partner", 1, { email: "a@x.com", phone: "1" });
    const email = { name: "email", type: "char" as const };
    const phone = { name: "phone", type: "char" as const, widget: "phone" };

    host.render(email, record, true);
    host.render(phone, record, true);
    host.invalidate("email");
    const again = host.render(email, record, true);
    expect((again.querySelector("input") as HTMLInputElement | null)?.value).toBe("a@x.com");
    host.clear();
  });
});
