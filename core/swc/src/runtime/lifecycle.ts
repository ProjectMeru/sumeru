import { getActiveHost, type HookHost, type StartFn } from "./hooks.js";

function requireActiveHost(): HookHost {
  const host = getActiveHost();
  if (!host) {
    throw new Error("Lifecycle hooks must run inside callSetup()");
  }
  return host;
}

/** Register async work that `runWillStart(host)` awaits before first paint. */
export function onWillStart(fn: StartFn): void {
  requireActiveHost().willStart.push(fn);
}

export function onWillPatch(fn: () => void): void {
  requireActiveHost().willPatch.push(fn);
}

export function onPatched(fn: () => void): void {
  requireActiveHost().patched.push(fn);
}

export async function runWillStart(host: HookHost): Promise<void> {
  for (const fn of host.willStart) {
    await fn();
  }
}

export function runWillPatch(host: HookHost): void {
  for (const fn of host.willPatch) {
    fn();
  }
}

export function runPatched(host: HookHost): void {
  for (const fn of host.patched) {
    fn();
  }
}

export { onMount, onWillUnmount } from "./hooks.js";
