type LifecycleFn = () => void;

const willPatchCallbacks: LifecycleFn[] = [];
const patchedCallbacks: LifecycleFn[] = [];
const willStartCallbacks: Array<() => void | Promise<void>> = [];
const willUpdatePropsCallbacks: Array<(next: unknown) => void | Promise<void>> = [];

let activeLifecycle: {
  willPatch: LifecycleFn[];
  patched: LifecycleFn[];
  willStart: Array<() => void | Promise<void>>;
  willUpdateProps: Array<(next: unknown) => void | Promise<void>>;
} | null = null;

function lifecycleTarget() {
  return activeLifecycle ?? {
    willPatch: willPatchCallbacks,
    patched: patchedCallbacks,
    willStart: willStartCallbacks,
    willUpdateProps: willUpdatePropsCallbacks,
  };
}

export function onWillPatch(fn: LifecycleFn): void {
  lifecycleTarget().willPatch.push(fn);
}

export function onPatched(fn: LifecycleFn): void {
  lifecycleTarget().patched.push(fn);
}

export function onWillStart(fn: () => void | Promise<void>): void {
  lifecycleTarget().willStart.push(fn);
}

export function onWillUpdateProps(fn: (next: unknown) => void | Promise<void>): void {
  lifecycleTarget().willUpdateProps.push(fn);
}

export async function runWillStart(): Promise<void> {
  for (const fn of lifecycleTarget().willStart.splice(0)) {
    await fn();
  }
}

export async function runWillUpdateProps(next: unknown): Promise<void> {
  for (const fn of lifecycleTarget().willUpdateProps.splice(0)) {
    await fn(next);
  }
}

export function runWillPatch(): void {
  for (const fn of lifecycleTarget().willPatch.splice(0)) {
    fn();
  }
}

export function runPatched(): void {
  for (const fn of lifecycleTarget().patched.splice(0)) {
    fn();
  }
}

export { onMount, onWillUnmount } from "./hooks.js";
