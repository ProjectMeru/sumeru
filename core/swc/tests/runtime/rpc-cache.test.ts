import { describe, expect, it, vi } from "vitest";
import { RpcService } from "../../src/services/rpc.js";

describe("RpcService searchRead cache", () => {
  it("dedupes identical search_read calls", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: [{ id: 1, name: "India" }] }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const rpc = new RpcService("/api/rpc", "tok", "test_db", "1");
    const a = rpc.searchRead("core.country", [], ["id", "name"], 200);
    const b = rpc.searchRead("core.country", [], ["id", "name"], 200);
    expect(a).toBe(b);
    await a;
    expect(fetchMock).toHaveBeenCalledTimes(1);

    vi.unstubAllGlobals();
  });

  it("clears cache on write", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      })
      .mockResolvedValueOnce({ ok: true, json: async () => ({ success: true, data: true }) })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: [{ id: 1 }] }),
      });
    vi.stubGlobal("fetch", fetchMock);

    const rpc = new RpcService("/api/rpc", "tok", "test_db", "1");
    await rpc.searchRead("core.country", [], ["id", "name"], 200);
    await rpc.write("core.partner", [1], { name: "Acme" });
    await rpc.searchRead("core.country", [], ["id", "name"], 200);
    expect(fetchMock).toHaveBeenCalledTimes(3);

    vi.unstubAllGlobals();
  });

  it("passes optional vals on call", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: { close: true } }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const rpc = new RpcService("/api/rpc", "tok", "test_db", "1");
    await rpc.callMethod("crm.lead", "action_merge_wizard", 3, { active_ids: "3,7" });
    const body = JSON.parse(fetchMock.mock.calls[0][1].body as string);
    expect(body).toEqual({ id: 3, method: "action_merge_wizard", vals: { active_ids: "3,7" } });

    vi.unstubAllGlobals();
  });
});
