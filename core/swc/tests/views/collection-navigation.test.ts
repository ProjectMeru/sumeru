import { describe, expect, it, vi } from "vitest";
import { openWorkspaceRecord, recordIdFromRow } from "../../src/views/shared/collection-navigation.js";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

function basePayload(): SwcWorkspacePayload {
  return {
    actionId: 12,
    menuId: "5",
    viewType: "list",
    model: "crm.lead",
    recordId: 0,
    formEdit: false,
    csrfToken: "tok",
    arch: { type: "list", model: "crm.lead", fields: [{ name: "name" }] },
    viewTabs: [],
    breadcrumbs: [],
  };
}

function testEnv(openRecord = vi.fn()): SwcEnv {
  return {
    services: { action: { openRecord } },
  } as unknown as SwcEnv;
}

describe("collection-navigation", () => {
  it("recordIdFromRow reads id from row or number", () => {
    expect(recordIdFromRow(42)).toBe(42);
    expect(recordIdFromRow({ id: 7 })).toBe(7);
    expect(recordIdFromRow({})).toBe(0);
  });

  it("openWorkspaceRecord calls action.openRecord with form view", () => {
    const openRecord = vi.fn();
    const env = testEnv(openRecord);
    const payload = basePayload();
    openWorkspaceRecord(env, payload, { id: 99 });
    expect(openRecord).toHaveBeenCalledWith({
      actionId: 12,
      menuId: "5",
      recordId: 99,
      viewType: "form",
    });
  });

  it("openWorkspaceRecord ignores invalid ids", () => {
    const openRecord = vi.fn();
    openWorkspaceRecord(testEnv(openRecord), basePayload(), { id: 0 });
    expect(openRecord).not.toHaveBeenCalled();
  });
});
