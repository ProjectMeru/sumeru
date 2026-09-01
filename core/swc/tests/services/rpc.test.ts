import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { RpcService } from "../../src/services/rpc.js";
import { SwcError } from "../../src/runtime/error.js";

describe("RpcService", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("searchRead caches identical requests", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true, result: [{ id: 1 }] }),
    } as Response);
    const rpc = new RpcService("/web/swc/rpc", "tok");
    const a = rpc.searchRead("m", [], ["name"], 10);
    const b = rpc.searchRead("m", [], ["name"], 10);
    expect(a).toBe(b);
    await a;
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it("write invalidates searchRead cache", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true, result: true }),
    } as Response);
    const rpc = new RpcService("/web/swc/rpc", "tok");
    await rpc.write("m", [1], { name: "x" });
    expect(fetch).toHaveBeenCalled();
  });

  it("throws SwcError on RPC failure envelope", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ ok: false, error: { message: "denied" } }),
    } as Response);
    const rpc = new RpcService("/web/swc/rpc", "tok");
    await expect(rpc.read("m", [1])).rejects.toBeInstanceOf(SwcError);
  });

  it("throws on HTTP error status", async () => {
    vi.mocked(fetch).mockResolvedValue({ ok: false, status: 502 } as Response);
    const rpc = new RpcService("/web/swc/rpc", "tok");
    await expect(rpc.read("m", [1])).rejects.toBeInstanceOf(SwcError);
  });

  it("create and unlink call dispatch", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true, result: 1 }),
    } as Response);
    const rpc = new RpcService("/web/swc/rpc", "tok");
    await rpc.create("m", { name: "a" });
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true, result: true }),
    } as Response);
    await rpc.unlink("m", [1]);
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it("callMethod, onchange, and readGroup dispatch", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true, result: { value: {} } }),
    } as Response);
    const rpc = new RpcService("/web/swc/rpc", "tok");
    await rpc.callMethod("m", "action_x", 3, { foo: "bar" });
    await rpc.onchange("m", { name: "a" }, "name");
    await rpc.readGroup("m", [], ["amount"], ["state"]);
    expect(fetch).toHaveBeenCalledTimes(3);
  });
});
