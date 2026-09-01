import { describe, expect, it, vi } from "vitest";
import { runObjectAction } from "../../src/views/shared/object-action.js";
import { SwcError } from "../../src/runtime/error.js";

function env(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    services: {
      rpc: { callMethod: vi.fn().mockResolvedValue({ close: true }) },
      action: { applyCallResult: vi.fn().mockResolvedValue(true) },
      notification: { success: vi.fn(), error: vi.fn() },
      ...((overrides.services as object) ?? {}),
    },
  } as never;
}

describe("runObjectAction", () => {
  it("returns true when action service handles result", async () => {
    const e = env();
    const handled = await runObjectAction(e, {
      model: "m",
      methodName: "action_done",
      recordId: 1,
      buttonLabel: "Done",
    });
    expect(handled).toBe(true);
    expect(e.services.rpc.callMethod).toHaveBeenCalledWith("m", "action_done", 1, undefined);
  });

  it("shows success notification when not handled", async () => {
    const e = env({ services: { action: { applyCallResult: vi.fn().mockResolvedValue(false) } } });
    const onSuccess = vi.fn();
    const handled = await runObjectAction(e, {
      model: "m",
      methodName: "refresh",
      recordId: 1,
      buttonLabel: "Refresh",
      onSuccess,
    });
    expect(handled).toBe(false);
    expect(e.services.notification.success).toHaveBeenCalled();
    expect(onSuccess).toHaveBeenCalled();
  });

  it("reports SwcError via notification", async () => {
    const e = env({
      services: {
        rpc: {
          callMethod: vi.fn().mockRejectedValue(new SwcError("blocked", "rpc_error")),
        },
      },
    });
    await runObjectAction(e, {
      model: "m",
      methodName: "x",
      recordId: 1,
      buttonLabel: "Run",
    });
    expect(e.services.notification.error).toHaveBeenCalledWith("Run", "blocked");
  });
});
