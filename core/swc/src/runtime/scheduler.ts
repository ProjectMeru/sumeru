/** Components waiting for a coalesced DOM patch. */
export interface ScheduledComponent {
  rootElement: HTMLElement | null;
  patch(): void;
}

const pending = new Set<ScheduledComponent>();
let frameHandle = 0;

function queueFrame(callback: () => void): number {
  if (typeof requestAnimationFrame === "function") {
    return requestAnimationFrame(callback);
  }
  return setTimeout(callback, 0) as unknown as number;
}

function cancelFrame(handle: number): void {
  if (typeof cancelAnimationFrame === "function") {
    cancelAnimationFrame(handle);
    return;
  }
  clearTimeout(handle);
}

/** Queue a connected component to patch on the next animation frame. */
export function scheduleRender(component: ScheduledComponent): void {
  pending.add(component);
  if (frameHandle) return;
  frameHandle = queueFrame(flushScheduledRenders);
}

/** Apply every queued patch immediately (tests and first-paint helpers). */
export function flushScheduledRenders(): void {
  if (frameHandle) {
    cancelFrame(frameHandle);
    frameHandle = 0;
  }
  const batch = [...pending];
  pending.clear();
  for (const component of batch) {
    if (component.rootElement?.isConnected) component.patch();
  }
}
