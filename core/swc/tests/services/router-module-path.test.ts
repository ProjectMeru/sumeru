import { describe, expect, it } from "vitest";
import { moduleNameFromWebPath, RouterService } from "../../src/services/router.js";

describe("moduleNameFromWebPath", () => {
  it("parses application module entry paths", () => {
    expect(moduleNameFromWebPath("/web/crm")).toBe("crm");
    expect(moduleNameFromWebPath("/web/engagement_cookbook")).toBe("engagement_cookbook");
  });

  it("rejects reserved and multi-segment paths", () => {
    expect(moduleNameFromWebPath("/web/apps")).toBe("apps");
    expect(moduleNameFromWebPath("/web/swc/workspace")).toBe("");
    expect(moduleNameFromWebPath("/web")).toBe("");
  });
});

describe("RouterService module pathname", () => {
  it("returns empty workspace route for module entry without query action", () => {
    const router = new RouterService();
    const route = router.parse({ pathname: "/web/crm", search: "" });
    expect(route.actionId).toBe(0);
    expect(route.menuId).toBe("");
  });
});
