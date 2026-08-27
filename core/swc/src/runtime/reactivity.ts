/** Reactive proxy layer (OWL-inspired, Sumeru-native). */

import { scheduleRender } from "./scheduler.js";

const RAW = Symbol("swc_raw");
const REACTIVE = Symbol("swc_reactive");

export function markRaw<T>(value: T): T {
  if (typeof value === "object" && value !== null) {
    Object.defineProperty(value, RAW, { value: true, enumerable: false });
  }
  return value;
}

export function toRaw<T>(value: T): T {
  if (typeof value === "object" && value !== null && (value as Record<symbol, unknown>)[RAW]) {
    return value;
  }
  return value;
}

type Subscriber = () => void;
const targetSubscribers = new WeakMap<object, Set<Subscriber>>();

function notify(target: object): void {
  targetSubscribers.get(target)?.forEach((fn) => fn());
}

export function reactive<T extends object>(target: T): T {
  if ((target as Record<symbol, unknown>)[REACTIVE]) return target;

  const proxy = new Proxy(target, {
    get(obj, key) {
      if (key === REACTIVE) return true;
      if (key === RAW) return obj;
      const val = Reflect.get(obj, key);
      if (typeof val === "object" && val !== null && !((val as Record<symbol, unknown>)[RAW])) {
        return reactive(val as object) as typeof val;
      }
      return val;
    },
    set(obj, key, value) {
      const result = Reflect.set(obj, key, value);
      notify(obj);
      return result;
    },
  });

  Object.defineProperty(proxy, REACTIVE, { value: true, enumerable: false });
  return proxy;
}

export function subscribe(target: object, fn: Subscriber): () => void {
  let set = targetSubscribers.get(target);
  if (!set) {
    set = new Set();
    targetSubscribers.set(target, set);
  }
  set.add(fn);
  return () => set!.delete(fn);
}

export function useReactiveState<T extends object>(initial: T): [T, () => void] {
  const state = reactive({ ...initial });
  subscribe(state, () => {
    scheduleRender();
  });
  return [state, () => scheduleRender()];
}
