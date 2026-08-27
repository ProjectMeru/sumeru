type EffectCleanup = () => void;
type EffectFn = () => void | EffectCleanup;

const mountCallbacks: EffectFn[] = [];
const unmountCallbacks: EffectFn[] = [];

export function runMountCallbacks(): void {
  for (const fn of mountCallbacks.splice(0)) {
    fn();
  }
}

export function runUnmountCallbacks(): void {
  for (const fn of unmountCallbacks.splice(0)) {
    fn();
  }
}

export function onMount(fn: EffectFn): void {
  mountCallbacks.push(fn);
}

export function onWillUnmount(fn: EffectFn): void {
  unmountCallbacks.push(fn);
}

export function useEffect(fn: EffectFn): void {
  onMount(() => {
    const cleanup = fn();
    if (typeof cleanup === "function") {
      onWillUnmount(cleanup);
    }
  });
}
