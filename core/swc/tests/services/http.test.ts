import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { stubFetch, unstubFetch } from "../harness/dom.js";
import { HttpService } from "../../src/services/http.js";
import { SwcError } from "../../src/runtime/error.js";

describe("HttpService", () => {
  beforeEach(() => {
    stubFetch();
  });

  afterEach(() => {
    unstubFetch();
  });

  it("getJSON returns parsed body on success", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true }),
    } as Response);
    const http = new HttpService("tok");
    await expect(http.getJSON<{ ok: boolean }>("/api/x")).resolves.toEqual({ ok: true });
  });

  it("getJSON throws SwcError on failure", async () => {
    vi.mocked(fetch).mockResolvedValue({ ok: false, status: 500 } as Response);
    const http = new HttpService("tok");
    await expect(http.getJSON("/api/x")).rejects.toBeInstanceOf(SwcError);
  });

  it("postJSON attaches csrf token", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ id: 1 }),
    } as Response);
    const http = new HttpService("csrf123");
    await http.postJSON("/api/save", { name: "a" });
    const [, init] = vi.mocked(fetch).mock.calls[0];
    expect(String(init?.body)).toContain("csrf123");
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf123");
  });

  it("delete throws on non-ok response", async () => {
    vi.mocked(fetch).mockResolvedValue({ ok: false, status: 403 } as Response);
    const http = new HttpService("tok");
    await expect(http.delete("/api/x")).rejects.toBeInstanceOf(SwcError);
  });

  it("postForm sends urlencoded body with csrf", async () => {
    vi.mocked(fetch).mockResolvedValue({ ok: true } as Response);
    const http = new HttpService("csrf99");
    await http.postForm("/upload", { model: "m" });
    const [, init] = vi.mocked(fetch).mock.calls[0];
    expect(String(init?.body)).toContain("csrf99");
  });

  it("csrf getter exposes token", () => {
    expect(new HttpService("my-token").csrf).toBe("my-token");
  });
});
