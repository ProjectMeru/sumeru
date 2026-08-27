import type { SwcServices } from "../runtime/env.js";
import { registry } from "../runtime/registry.js";

/** Register built-in service instances on the global registry for addon discovery. */
export function registerCoreServices(services: SwcServices): void {
  const cat = registry.category("services");
  for (const [key, instance] of Object.entries(services)) {
    cat.add(key, instance);
  }
}
