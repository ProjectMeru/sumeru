import { describe, expect, it } from "vitest";
import {
  isReloadNavigation,
  pruneStaleTabs,
  registerTab,
  registryIsEmpty,
  unregisterTab,
} from "../../src/shell/shell-tab-presence.js";

describe("shell-tab-presence registry", () => {
  it("registers and unregisters tabs", () => {
    let reg = registerTab({}, "a", 100);
    reg = registerTab(reg, "b", 200);
    expect(registryIsEmpty(reg)).toBe(false);
    reg = unregisterTab(reg, "a");
    expect(reg).toEqual({ b: 200 });
    reg = unregisterTab(reg, "b");
    expect(registryIsEmpty(reg)).toBe(true);
  });

  it("prunes stale tab entries", () => {
    const reg = { old: 100, fresh: 5000 };
    const pruned = pruneStaleTabs(reg, 1000, 6000);
    expect(pruned).toEqual({ fresh: 5000 });
  });

  it("detects reload navigation type", () => {
    expect(typeof isReloadNavigation()).toBe("boolean");
  });
});
