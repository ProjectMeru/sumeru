import { vi } from "vitest";

/** Create a DOM element with optional class, text, and attributes. */
export function createElement<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  init?: { className?: string; text?: string; attrs?: Record<string, string> },
): HTMLElementTagNameMap[K] {
  const el = document.createElement(tag);
  if (init?.className) el.className = init.className;
  if (init?.text) el.textContent = init.text;
  if (init?.attrs) {
    for (const [key, value] of Object.entries(init.attrs)) {
      el.setAttribute(key, value);
    }
  }
  return el;
}

/** Stub global fetch for service tests; call unstubFetch in afterEach. */
export function stubFetch(): typeof fetch {
  vi.stubGlobal("fetch", vi.fn());
  return vi.mocked(fetch);
}

export function unstubFetch(): void {
  vi.unstubAllGlobals();
}
