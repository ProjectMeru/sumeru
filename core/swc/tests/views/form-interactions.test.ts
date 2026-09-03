import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { createElement } from "../harness/dom.js";
import { initFormInteractions } from "../../src/views/form/form-interactions.js";

describe("form-interactions", () => {
  let root: HTMLElement;
  let cleanup: (() => void) | undefined;

  beforeEach(() => {
    root = createElement("div");
    root.innerHTML = `
      <div class="sum-notebook-tabs">
        <button type="button" role="tab">A</button>
        <button type="button" role="tab">B</button>
      </div>
      <div class="sum-field-widget sum-field-widget--many2one">
        <div class="sum-m2o-suggest"></div>
      </div>
      <details class="sum-date-field" open>
        <summary>Date</summary>
      </details>
      <div data-sum-avatar data-sum-readonly="0">
        <div class="sum-form-avatar-box sum-form-avatar-box--clickable" tabindex="0">
          <span class="sum-form-avatar-initials">AB</span>
        </div>
        <input type="file" class="sum-image-file-input" />
        <input type="hidden" data-sum-avatar-value="" />
      </div>
    `;
    document.body.append(root);
    cleanup = initFormInteractions(root);
  });

  afterEach(() => {
    cleanup?.();
    document.body.innerHTML = "";
  });

  it("arrow keys move notebook tab focus", () => {
    const tabs = root.querySelectorAll<HTMLButtonElement>('button[role="tab"]');
    tabs[0].focus();
    tabs[0].dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }));
    expect(document.activeElement).toBe(tabs[1]);
  });

  it("document click dismisses many2one suggestions", () => {
    expect(root.querySelector(".sum-m2o-suggest")).toBeTruthy();
    document.body.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    expect(root.querySelector(".sum-m2o-suggest")).toBeNull();
  });

  it("document click closes open date details", () => {
    const details = root.querySelector<HTMLDetailsElement>("details.sum-date-field")!;
    expect(details.open).toBe(true);
    document.body.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    expect(details.open).toBe(false);
  });

  it("escape removes many2one suggestion lists", () => {
    root.querySelector(".sum-field-widget--many2one")!.insertAdjacentHTML(
      "beforeend",
      '<div class="sum-m2o-suggest"></div>',
    );
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    expect(root.querySelector(".sum-m2o-suggest")).toBeNull();
  });
});
