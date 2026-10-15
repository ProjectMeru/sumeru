import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { RpcService } from "../../src/services/rpc.js";
import { SwcError } from "../../src/runtime/error.js";
import { resetSessionRedirectForTests } from "../../src/services/session-redirect.js";

function testRpc(): RpcService {
  return new RpcService("/api/rpc", "tok", "test_db", "1");
}

describe("RpcService", () => {
  beforeEach(() => {
    resetSessionRedirectForTests();
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("searchRead caches identical requests", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: [{ id: 1 }] }),
    } as Response);
    const rpc = testRpc();
    const a = rpc.searchRead("m", [], ["name"], 10);
    const b = rpc.searchRead("m", [], ["name"], 10);
    expect(a).toBe(b);
    await a;
    expect(fetch).toHaveBeenCalledTimes(1);
    const [url] = vi.mocked(fetch).mock.calls[0] as [string];
    expect(url).toBe("/api/rpc/m/search_read");
  });

  it("write invalidates searchRead cache", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: true }),
    } as Response);
    const rpc = testRpc();
    await rpc.write("m", [1], { name: "x" });
    expect(fetch).toHaveBeenCalled();
  });

  it("throws SwcError on RPC failure envelope", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ success: false, error: { message: "denied", context: { code: "ACCESS_DENIED" } } }),
    } as Response);
    const rpc = testRpc();
    await expect(rpc.read("m", [1])).rejects.toBeInstanceOf(SwcError);
  });

  it("throws on HTTP error status", async () => {
    vi.mocked(fetch).mockResolvedValue({ ok: false, status: 502 } as Response);
    const rpc = testRpc();
    await expect(rpc.read("m", [1])).rejects.toBeInstanceOf(SwcError);
  });

  it("redirects to login on HTTP 401", async () => {
    const assign = vi.fn();
    vi.stubGlobal("location", { pathname: "/web/home", search: "", assign });
    vi.mocked(fetch).mockResolvedValue({ ok: false, status: 401 } as Response);
    const rpc = testRpc();
    await expect(rpc.read("m", [1])).rejects.toBeInstanceOf(SwcError);
    expect(assign).toHaveBeenCalledWith("/web/login?next=%2Fweb%2Fhome");
  });

  it("create and unlink call dispatch", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: 1 }),
    } as Response);
    const rpc = testRpc();
    await rpc.create("m", { name: "a" });
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: true }),
    } as Response);
    await rpc.unlink("m", [1]);
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it("callMethod, onchange, and readGroup dispatch", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: { value: {} } }),
    } as Response);
    const rpc = testRpc();
    await rpc.callMethod("m", "action_x", 3, { foo: "bar" });
    await rpc.onchange("m", { name: "a" }, "name");
    await rpc.readGroup("m", [], ["amount"], ["state"]);
    expect(fetch).toHaveBeenCalledTimes(3);
  });
});
