import type { SwcServices } from "./env.js";
import { scheduleRender } from "./scheduler.js";

export type EffectCleanup = () => void;
export type EffectFn = () => void | EffectCleanup;
export type StartFn = () => void | Promise<void>;

/** Host that hook registrars attach to during `callSetup()`. */
export interface HookHost {
  env: { services: SwcServices };
  rootElement: HTMLElement | null;
  hookState: unknown[];
  hookIndex: number;
  willStart: StartFn[];
  willPatch: Array<() => void>;
  patched: Array<() => void>;
  mountEffects: EffectFn[];
  unmountEffects: EffectFn[];
  consumeHookSlot(): number;
  patch(): void;
}

let activeHost: HookHost | null = null;

export function getActiveHost(): HookHost | null {
  return activeHost;
}

function requireActiveHost(): HookHost {
  if (!activeHost) {
    throw new Error("Hooks must run inside callSetup()");
  }
  return activeHost;
}

export function withActiveHost<T>(host: HookHost, fn: () => T): T {
  const previous = activeHost;
  activeHost = host;
  try {
    return fn();
  } finally {
    activeHost = previous;
  }
}

export function onMount(fn: EffectFn): void {
  requireActiveHost().mountEffects.push(fn);
}

export function onWillUnmount(fn: EffectFn): void {
  requireActiveHost().unmountEffects.push(fn);
}

export function useEffect(fn: EffectFn): void {
  onMount(() => {
    const cleanup = fn();
    if (typeof cleanup === "function") {
      onWillUnmount(cleanup);
    }
  });
}

export interface StateBox<T> {
  readonly value: T;
}

export function useState<T>(
  initial: T,
): [StateBox<T>, (next: T | ((previous: T) => T)) => void] {
  const host = requireActiveHost();
  const index = host.consumeHookSlot();
  if (host.hookState[index] === undefined) {
    host.hookState[index] = initial;
  }
  const box: StateBox<T> = {
    get value() {
      return host.hookState[index] as T;
    },
  };
  const setValue = (next: T | ((previous: T) => T)): void => {
    const previous = host.hookState[index] as T;
    const value = typeof next === "function" ? (next as (previous: T) => T)(previous) : next;
    if (Object.is(value, previous)) return;
    host.hookState[index] = value;
    scheduleRender(host);
  };
  return [box, setValue];
}

export function useService<K extends keyof SwcServices>(name: K): SwcServices[K] {
  return requireActiveHost().env.services[name];
}

export function runMountEffects(host: HookHost): void {
  for (const fn of host.mountEffects) {
    fn();
  }
}

export function runUnmountEffects(host: HookHost): void {
  for (const fn of host.unmountEffects) {
    fn();
  }
}

function cssAttrEscape(name: string): string {
  return name.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}

/** Query a `data-ref` node under the active component root (after mount/patch). */
export function useTemplateRef(name: string): { current: Element | null } {
  const host = requireActiveHost();
  const index = host.consumeHookSlot();
  if (host.hookState[index] === undefined) {
    host.hookState[index] = { current: null as Element | null };
  }
  const box = host.hookState[index] as { current: Element | null };
  const refresh = (): void => {
    box.current = host.rootElement?.querySelector(`[data-ref="${cssAttrEscape(name)}"]`) ?? null;
  };
  onMount(refresh);
  host.patched.push(refresh);
  return box;
}
