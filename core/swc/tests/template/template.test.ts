import { describe, expect, it, vi } from "vitest";
import { html } from "../../src/template/html.js";

describe("html template", () => {
  it("nests text inside elements with closing tags", () => {
    const el = html`<div class="sum-workspace-loading">Loading workspace…</div>`.render();
    expect(el.tagName).toBe("DIV");
    expect(el.className).toBe("sum-workspace-loading");
    expect(el.textContent).toBe("Loading workspace…");
  });

  it("renders nested structure like ErrorBoundary", () => {
    const el = html`
      <div class="sum-flash sum-flash--error">
        <strong>SWC error</strong>
        <p>something broke</p>
      </div>
    `.render();
    expect(el.querySelector("strong")?.textContent).toBe("SWC error");
    expect(el.querySelector("p")?.textContent).toBe("something broke");
  });

  it("renders table rows with nested cells", () => {
    const el = html`<table><tr><td>cell</td></tr></table>`.render();
    expect(el.querySelector("td")?.textContent).toBe("cell");
  });

  it("interpolates href attributes split across values", () => {
    const url = "/web?action=12&view_type=form";
    const el = html`<a class="sum-btn" href=${url}>New</a>`.render();
    expect(el.tagName).toBe("A");
    expect(el.getAttribute("href")).toBe(url);
    expect(el.textContent).toBe("New");
  });

  it("binds click handlers on interpolated attributes", () => {
    const onClick = vi.fn();
    const el = html`<button type="button" @click=${onClick}>Save</button>`.render();
    el.click();
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("embeds HTMLElement children", () => {
    const link = document.createElement("a");
    link.href = "/web/export/csv";
    link.textContent = "Export CSV";
    const el = html`<div class="sum-view-toolbar-actions">${link}</div>`.render();
    expect(el.querySelector("a")?.textContent).toBe("Export CSV");
  });

  it("parses quoted class attributes with spaces", () => {
    const el = html`<div class="sum-form-split-layout sum-form-split-layout--compact">Title</div>`.render();
    expect(el.className).toBe("sum-form-split-layout sum-form-split-layout--compact");
  });

  it("interpolates class values containing spaces", () => {
    const tabClass = "sum-notebook-tab sum-notebook-tab--active";
    const el = html`<button type="button" class=${tabClass}>Notes</button>`.render();
    expect(el.className).toBe("sum-notebook-tab sum-notebook-tab--active");
  });

  it("quotes placeholder values with trailing dots", () => {
    const ph = "Opportunity...";
    const el = html`<input class="sum-form-hero-input" placeholder=${ph} value="" />`.render();
    expect(el.getAttribute("placeholder")).toBe("Opportunity...");
  });

  it("quotes empty interpolated attribute values", () => {
    const val = "";
    const el = html`<input class="sum-field-input" placeholder="Name" value=${val} />`.render();
    expect(el.getAttribute("value")).toBe("");
    expect(el.getAttribute("placeholder")).toBe("Name");
  });

  it("quotes hero title input attributes across lines", () => {
    const name = "name";
    const placeholder = "Opportunity...";
    const val = "";
    const el = html`<h1>
      <input
        class="sum-form-hero-input sum-form-hero-input--bold"
        name=${name}
        placeholder=${placeholder}
        value=${val}
        aria-label=${placeholder}
      />
    </h1>`.render();
    const input = el.querySelector(".sum-form-hero-input") as HTMLInputElement;
    expect(input?.getAttribute("placeholder")).toBe("Opportunity...");
    expect(input?.getAttribute("value")).toBe("");
  });

  it("omits optional attributes when interpolation is undefined", () => {
    const el = html`<button type="button" disabled=${undefined}>Save</button>`.render();
    expect(el.hasAttribute("disabled")).toBe(false);
  });

  it("ignores empty TemplateResult suffix in tag attributes", () => {
    const suffix = html``;
    const el = html`<label for=${"f-date_start"}${suffix}>Start</label>`.render();
    expect(el.getAttribute("for")).toBe("f-date_start");
  });
});
