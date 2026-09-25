import { vi } from "vitest";
import type { SwcEnv } from "../../src/runtime/env.js";
import type { SwcViewArch, SwcWorkspacePayload } from "../../src/types/workspace.js";
import { RecordService } from "../../src/model/record.js";
import { CommandService } from "../../src/services/command.js";
import { BusService } from "../../src/services/bus.js";

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
  const bus = (services.bus as BusService | undefined) ?? new BusService();
  const rpc =
    services.rpc ??
    ({
      write: vi.fn().mockResolvedValue(undefined),
      create: vi.fn().mockResolvedValue(99),
      unlink: vi.fn().mockResolvedValue(undefined),
      read: vi.fn().mockResolvedValue([{ id: 1, name: "Test" }]),
      searchRead: vi.fn().mockResolvedValue([]),
      call: vi.fn().mockResolvedValue(undefined),
      onchange: vi.fn().mockResolvedValue({}),
    } as SwcEnv["services"]["rpc"]);
  return {
    bootstrap: { swcApiBase: "/web/swc" } as never,
    services: {
      action: { openRecord: vi.fn(), navigate: vi.fn() },
      router: { workspaceUrl: () => "/web", parse: vi.fn() },
      rpc,
      http: {
        getJSON: vi.fn().mockResolvedValue({ messages: [], enabled: true }),
        postForm: vi.fn().mockResolvedValue(undefined),
        postJSON: vi.fn().mockResolvedValue({}),
      },
      notification: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
      dialog: { confirm: vi.fn().mockResolvedValue(true) },
      bus,
      record: services.record ?? new RecordService(rpc, bus),
      command: services.command ?? new CommandService(),
      ...services,
      bus,
      rpc,
    },
  } as unknown as SwcEnv;
}
