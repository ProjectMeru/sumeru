import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { flushScheduledRenders, scheduleRender } from "../../src/runtime/scheduler.js";

describe("scheduler coalesced renders", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "requestAnimationFrame",
      (cb: FrameRequestCallback) => setTimeout(() => cb(0), 0) as unknown as number,
    );
    vi.stubGlobal("cancelAnimationFrame", (id: number) => clearTimeout(id));
  });

  afterEach(() => {
    flushScheduledRenders();
    vi.unstubAllGlobals();
  });

  it("batches multiple scheduleRender calls into one flush", async () => {
    let patches = 0;
    const rootElement = document.createElement("div");
    document.body.append(rootElement);
    const component = {
      rootElement,
      patch() {
        patches += 1;
      },
    };
    scheduleRender(component);
    scheduleRender(component);
    expect(patches).toBe(0);
    flushScheduledRenders();
    expect(patches).toBe(1);
  });
});
