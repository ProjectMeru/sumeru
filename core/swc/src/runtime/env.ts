import type { SwcBootstrap } from "../types/bootstrap.js";
import type { RpcService } from "../services/rpc.js";
import type { HttpService } from "../services/http.js";
import type { NotificationService } from "../services/notification.js";
import type { ActionService } from "../services/action.js";
import type { RouterService } from "../services/router.js";
import type { BusService } from "../services/bus.js";
import type { DialogService } from "../services/dialog.js";

export interface SwcServices {
  rpc: RpcService;
  http: HttpService;
  notification: NotificationService;
  action: ActionService;
  router: RouterService;
  bus: BusService;
  dialog: DialogService;
}

export class SwcEnv {
  readonly bootstrap: SwcBootstrap;
  readonly services: SwcServices;

  constructor(bootstrap: SwcBootstrap, services: SwcServices) {
    this.bootstrap = bootstrap;
    this.services = services;
  }

  get<T extends keyof SwcServices>(name: T): SwcServices[T] {
    return this.services[name];
  }
}
