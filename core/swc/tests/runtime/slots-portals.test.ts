import { describe, expect, it } from "vitest";
import { html } from "../../src/template/html.js";
import { SwcComponent } from "../../src/runtime/component.js";
import { mountComponent } from "../../src/runtime/component-host.js";
import { applyPortals, restorePortals } from "../../src/runtime/portals.js";
import { useTemplateRef } from "../../src/runtime/hooks.js";
import type { SwcEnv } from "../../src/runtime/env.js";

function testEnv(): SwcEnv {
  return { bootstrap: {} as never, services: {} } as unknown as SwcEnv;
}

class SlotTarget extends SwcComponent {
  override template() {
    return html`<div class="sum-slot-root"><div data-slot="body"></div></div>`;
  }
}

class RefHost extends SwcComponent {
  input!: { current: Element | null };

  override setup(): void {
    this.input = useTemplateRef("title");
  }

  override template() {
    return html`<div><input data-ref="title" value="ok" /></div>`;
  }
}

describe("slots and portals", () => {
  it("fills data-slot targets from ComponentHost", () => {
    const host = mountComponent(SlotTarget, {}, testEnv(), {
      body: html`<span class="sum-slot-body">Hello</span>`,
    });
    const el = host.render();
    expect(el.querySelector(".sum-slot-body")?.textContent).toBe("Hello");
    host.destroy();
  });

  it("moves data-portal nodes and restores them", () => {
    const root = document.createElement("div");
    const node = document.createElement("div");
    node.dataset.portal = "body";
    node.className = "sum-portal-node";
    root.append(node);
    document.body.append(root);
    applyPortals(root);
    expect(node.parentElement).toBe(document.body);
    restorePortals(root);
    expect(node.parentElement).toBe(root);
    root.remove();
  });

  it("resolves data-ref after mount", () => {
    const view = new RefHost({}, testEnv());
    view.callSetup();
    document.body.append(view.render());
    expect(view.input.current).toBeInstanceOf(HTMLInputElement);
    view.destroy();
  });
});
