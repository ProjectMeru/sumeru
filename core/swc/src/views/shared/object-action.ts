import type { SwcEnv } from "../../runtime/env.js";
import { SwcError } from "../../runtime/error.js";

export interface RunObjectActionOptions {
  model: string;
  methodName: string;
  recordId: number;
  extraArgs?: Record<string, string>;
  buttonLabel: string;
  onSuccess?: () => void | Promise<void>;
}

/**
 * Run a registered object method via RPC, then apply close/open/redirect.
 * Returns true when navigation or a dialog consumed the result.
 */
export async function runObjectAction(
  env: SwcEnv,
  options: RunObjectActionOptions,
): Promise<boolean> {
  try {
    const result = await env.services.rpc.callMethod(
      options.model,
      options.methodName,
      options.recordId,
      options.extraArgs,
    );
    if (await env.services.action.applyCallResult(result)) {
      return true;
    }
    if (options.methodName === "action_set_won") {
      env.services.notification.success("Congratulations!", "Another win!");
    } else {
      env.services.notification.success(options.buttonLabel, "Action completed.");
    }
    await options.onSuccess?.();
    return false;
  } catch (error) {
    env.services.notification.error(
      options.buttonLabel,
      error instanceof SwcError ? error.message : String(error),
    );
    return false;
  }
}
