import { vi } from "vitest";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcViewArch, SwcWorkspacePayload } from "../../src/types/workspace.js";

export function viewPayload(
  arch: Partial<SwcViewArch>,
  records: Record<string, unknown>[] = [],
): SwcWorkspacePayload {
  return {
    actionId: 1,
    menuId: "2",
    viewType: arch.type ?? "list",
    model: "demo.model",
    recordId: 0,
    formEdit: false,
    csrfToken: "tok",
    arch: {
      type: arch.type ?? "list",
      model: "demo.model",
      fields: arch.fields ?? [{ name: "name" }],
      ...arch,
    },
    records,
    viewTabs: [],
    breadcrumbs: [],
  };
}

export function collectionEnv(services: Partial<SwcEnv["services"]> = {}): SwcEnv {
  return {
    bootstrap: { swcApiBase: "/web/swc" } as never,
    services: {
      action: { openRecord: vi.fn(), navigate: vi.fn() },
      router: { workspaceUrl: () => "/web", parse: vi.fn() },
      rpc: {
        write: vi.fn().mockResolvedValue(undefined),
        create: vi.fn().mockResolvedValue(99),
        unlink: vi.fn().mockResolvedValue(undefined),
        read: vi.fn().mockResolvedValue([{ id: 1, name: "Test" }]),
        searchRead: vi.fn().mockResolvedValue([]),
        call: vi.fn().mockResolvedValue(undefined),
      },
      http: {
        getJSON: vi.fn().mockResolvedValue({ messages: [], enabled: true }),
        postForm: vi.fn().mockResolvedValue(undefined),
        postJSON: vi.fn().mockResolvedValue({}),
      },
      notification: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
      dialog: { confirm: vi.fn().mockResolvedValue(true) },
      bus: { emit: vi.fn(), subscribe: vi.fn() },
      ...services,
    },
  } as unknown as SwcEnv;
}
