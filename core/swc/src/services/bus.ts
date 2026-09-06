import { SWC_API_BASE } from "../constants/routes.js";

type BusHandler = (payload: unknown) => void;

function parseBusMessage(raw: unknown): { channel: string; payload: unknown } | null {
  if (typeof raw !== "object" || raw === null) return null;
  const channel = (raw as { channel?: unknown }).channel;
  if (typeof channel !== "string" || channel === "") return null;
  return { channel, payload: (raw as { payload?: unknown }).payload };
}

/** Client event bus with optional WebSocket live updates from /web/swc/bus. */
export class BusService {
  private readonly handlers = new Map<string, Set<BusHandler>>();
  private ws: WebSocket | null = null;

  subscribe(channel: string, handler: BusHandler): () => void {
    if (!this.handlers.has(channel)) {
      this.handlers.set(channel, new Set());
    }
    this.handlers.get(channel)!.add(handler);
    return () => this.handlers.get(channel)?.delete(handler);
  }

  emit(channel: string, payload: unknown): void {
    for (const fn of this.handlers.get(channel) ?? []) {
      fn(payload);
    }
  }

  /** Connect to /web/swc/bus when bootstrap.busEnabled is true. */
  connect(url = `${SWC_API_BASE}/bus`): void {
    if (this.ws) return;
    try {
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      this.ws = new WebSocket(`${proto}//${window.location.host}${url}`);
      this.ws.addEventListener("message", (ev) => {
        try {
          const parsed: unknown = JSON.parse(String(ev.data));
          const msg = parseBusMessage(parsed);
          if (msg) this.emit(msg.channel, msg.payload);
        } catch (err) {
          console.warn("swc bus: malformed message", err);
        }
      });
      this.ws.addEventListener("close", () => {
        this.ws = null;
      });
    } catch (err) {
      console.warn("swc bus: WebSocket unavailable; local-only bus", err);
    }
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
}
