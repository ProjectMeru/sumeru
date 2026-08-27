import { describe, expect, it } from "vitest";
import { html } from "../../src/template/html.js";
import { forEach } from "../../src/template/helpers.js";

describe("html patch", () => {
  it("updates attributes in place", () => {
    const first = html`<div class="sum-a" title="one">Hello</div>`.render();
    const patched = html`<div class="sum-b" title="two">Hello</div>`.patch(first);
    expect(patched).toBe(first);
    expect(patched.className).toBe("sum-b");
    expect(patched.getAttribute("title")).toBe("two");
  });

  it("reorders keyed children without recreating them", () => {
    const items = [
      { id: 1, name: "A" },
      { id: 2, name: "B" },
    ];
    const list = html`<ul>${forEach(items, (item) => item.id, (item) => html`<li>${item.name}</li>`)}</ul>`.render();
    const first = list.children[0] as HTMLElement;
    const second = list.children[1] as HTMLElement;
    const reversed = [...items].reverse();
    html`<ul>${forEach(reversed, (item) => item.id, (item) => html`<li>${item.name}</li>`)}</ul>`.patch(list);
    expect([...list.children]).toEqual([second, first]);
    expect(second.textContent).toBe("B");
    expect(first.textContent).toBe("A");
  });

  it("preserves input focus across a patch", () => {
    document.body.append(
      html`<div class="sum-wrap"><input class="sum-field-input" value="before" /></div>`.render(),
    );
    const wrap = document.querySelector(".sum-wrap") as HTMLElement;
    const input = wrap.querySelector("input") as HTMLInputElement;
    input.focus();
    expect(document.activeElement).toBe(input);
    html`<div class="sum-wrap sum-wrap--dirty"><input class="sum-field-input" value="after" /></div>`.patch(wrap);
    expect(document.activeElement).toBe(input);
    expect(wrap.className).toBe("sum-wrap sum-wrap--dirty");
    wrap.remove();
  });
});
