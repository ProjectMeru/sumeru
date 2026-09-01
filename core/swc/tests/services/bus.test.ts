import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { BusService } from "../../src/services/bus.js";

describe("BusService", () => {
  let bus: BusService;

  beforeEach(() => {
    bus = new BusService();
  });

  it("subscribe and emit deliver payloads", () => {
    const received: unknown[] = [];
    const unsub = bus.subscribe("ch", (p) => received.push(p));
    bus.emit("ch", { x: 1 });
    expect(received).toEqual([{ x: 1 }]);
    unsub();
    bus.emit("ch", { x: 2 });
    expect(received).toEqual([{ x: 1 }]);
  });

  it("disconnect clears websocket", () => {
    const close = vi.fn();
    (bus as unknown as { ws: WebSocket | null }).ws = { close } as unknown as WebSocket;
    bus.disconnect();
    expect(close).toHaveBeenCalled();
    expect((bus as unknown as { ws: WebSocket | null }).ws).toBeNull();
  });

  it("connect is no-op when already connected", () => {
    (bus as unknown as { ws: WebSocket | null }).ws = { close: vi.fn() } as unknown as WebSocket;
    const WS = vi.fn();
    vi.stubGlobal("WebSocket", WS);
    bus.connect("/web/swc/bus");
    expect(WS).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });

  afterEach(() => {
    bus.disconnect();
  });
});
