import { describe, expect, it } from "vitest";
import { SwcComponent } from "../../src/runtime/component.js";
import { html } from "../../src/template/html.js";
import { useService, useState, type StateBox } from "../../src/runtime/hooks.js";
import { onWillStart, runWillStart } from "../../src/runtime/lifecycle.js";
import { flushScheduledRenders } from "../../src/runtime/scheduler.js";
import { DialogService } from "../../src/services/dialog.js";
import type { SwcEnv } from "../../src/runtime/env.js";

function testEnv(dialog = new DialogService()): SwcEnv {
  return {
    bootstrap: {} as never,
    services: { dialog },
  } as unknown as SwcEnv;
}

class CounterView extends SwcComponent {
  count!: StateBox<number>;
  setCount!: (next: number | ((previous: number) => number)) => void;

  override setup(): void {
    const [count, setCount] = useState(0);
    this.count = count;
    this.setCount = setCount;
  }

  override template() {
    return html`<span class="sum-counter">${this.count.value}</span>`;
  }
}

class StartView extends SwcComponent {
  started = 0;

  override setup(): void {
    onWillStart(() => {
      this.started += 1;
    });
  }

  override template() {
    return html`<div class="sum-start"></div>`;
  }
}

class ServiceView extends SwcComponent {
  dialog!: DialogService;

  override setup(): void {
    this.dialog = useService("dialog");
  }

  override template() {
    return html`<div></div>`;
  }
}

describe("useState", () => {
  it("batches setValue into one patch", () => {
    const view = new CounterView({}, testEnv());
    view.callSetup();
    document.body.append(view.render());
    view.setCount(1);
    view.setCount(2);
    expect(view.rootElement?.textContent).toBe("0");
    flushScheduledRenders();
    expect(view.rootElement?.textContent).toBe("2");
    view.destroy();
  });
});

describe("onWillStart", () => {
  it("runs only on the instance that registered it", async () => {
    const first = new StartView({}, testEnv());
    const second = new StartView({}, testEnv());
    first.callSetup();
    second.callSetup();
    await runWillStart(first);
    expect(first.started).toBe(1);
    expect(second.started).toBe(0);
  });
});

describe("useService", () => {
  it("reads the named service from the component env", () => {
    const dialog = new DialogService();
    const view = new ServiceView({}, testEnv(dialog));
    view.callSetup();
    expect(view.dialog).toBe(dialog);
  });
});
