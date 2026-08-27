type StartFn = () => void | Promise<void>;

const willStartCallbacks: StartFn[] = [];

/** Register async setup that SwcApp.mount and WorkspaceRouter await before first paint. */
export function onWillStart(fn: StartFn): void {
  willStartCallbacks.push(fn);
}

export async function runWillStart(): Promise<void> {
  for (const fn of willStartCallbacks.splice(0)) {
    await fn();
  }
}

export function runWillPatch(): void {
  // Reserved for per-patch hooks; no registrars in core yet.
}

export function runPatched(): void {
  // Reserved for per-patch hooks; no registrars in core yet.
}

export { onMount, onWillUnmount } from "./hooks.js";
