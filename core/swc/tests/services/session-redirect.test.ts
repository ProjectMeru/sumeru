import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import {
  redirectToLoginIfNeeded,
  resetSessionRedirectForTests,
  shouldRedirectForHttpStatus,
  shouldRedirectForRpcError,
} from "../../src/services/session-redirect.js";

describe("session-redirect", () => {
  beforeEach(() => {
    resetSessionRedirectForTests();
    vi.stubGlobal("location", { pathname: "/web/home", search: "", assign: vi.fn() });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("detects unauthorized HTTP and RPC codes", () => {
    expect(shouldRedirectForHttpStatus(401)).toBe(true);
    expect(shouldRedirectForHttpStatus(403)).toBe(false);
    expect(shouldRedirectForRpcError("UNAUTHORIZED")).toBe(true);
    expect(shouldRedirectForRpcError("ACCESS_DENIED")).toBe(false);
  });

  it("redirects to login with next path", () => {
    const assign = vi.fn();
    vi.stubGlobal("location", {
      pathname: "/web/home",
      search: "?menu_id=1",
      assign,
    });
    redirectToLoginIfNeeded();
    expect(assign).toHaveBeenCalledWith("/web/login?next=%2Fweb%2Fhome%3Fmenu_id%3D1");
  });

  it("skips redirect on login route", () => {
    const assign = vi.fn();
    vi.stubGlobal("location", { pathname: "/web/login", search: "", assign });
    redirectToLoginIfNeeded();
    expect(assign).not.toHaveBeenCalled();
  });
});
