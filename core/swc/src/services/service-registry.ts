import type { SwcServices } from "../runtime/env.js";
import { registry } from "../runtime/registry.js";

/** Register built-in services on the global registry for addon discovery. */
export function registerCoreServices(services: SwcServices): void {
  const cat = registry.category("services");
  for (const [key, svc] of Object.entries(services)) {
    cat.add(key, svc as never);
  }
}

export function getService<K extends keyof SwcServices>(
  env: { services: SwcServices },
  name: K,
): SwcServices[K] {
  return env.services[name];
}
