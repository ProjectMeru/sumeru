import { describe, expect, it } from "vitest";
import { renderFormSheet } from "../../src/views/form/form-sheet.js";
import { SwcRecord } from "../../src/model/record.js";
import { registerDefaultWidgets } from "../../src/widgets/registry.js";
import type { SwcArchSheet } from "../../src/types/workspace.js";

registerDefaultWidgets();

function partnerSheet(): SwcArchSheet {
  return {
    divs: [{ class: "sum_title", h1Fields: [{ name: "name", string: "Name", placeholder: "Contact or Company Name..." }] }],
    fields: [],
    groups: [
      {
        fields: [],
        groups: [
          { string: "Contact", fields: [{ name: "email", string: "Email", type: "char" }] },
          { string: "Address", fields: [{ name: "street", string: "Street", type: "char" }] },
        ],
      },
    ],
    notebook: [
      {
        pages: [{ title: "Notes", fields: [{ name: "comment", string: "Notes", type: "text" }], groups: [] }],
      },
    ],
  };
}

describe("form-sheet", () => {
  it("renders title, groups, notebook, and split avatar layout", () => {
    const record = new SwcRecord("core.partner", 1, { name: "Acme", email: "a@x.com" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: partnerSheet(),
      record,
      readonly: true,
      hasImageField: true,
      activeNotebookPages: { 0: 0 },
      onNotebookTab: () => {},
    }).render();

    expect(el.matches(".sum-form-sheet") || el.querySelector(".sum-form-sheet")).toBeTruthy();
    expect(el.querySelector(".sum-form-split-layout")).toBeTruthy();
    expect(el.querySelector(".sum-form-avatar-img")?.getAttribute("src")).toContain(
      "/static/img/image_placeholder.jpg",
    );
    expect(el.textContent).toContain("Acme");
    expect(el.textContent).toContain("Contact");
    expect(el.textContent).toContain("Notes");
    expect((el.querySelector('input[name="email"]') as HTMLInputElement | null)?.value).toBe("a@x.com");
    expect(el.querySelector(".sum-form-group--col .sum-form-group-title")?.textContent).toBe("Contact");
    expect(el.querySelector(".sum-form-group-row")).toBeTruthy();
    expect(el.querySelector(".sum-form-group-span")).toBeTruthy();
    expect(el.querySelector(".sum-form-edit-grid > .sum-form-group--full")).toBeNull();
    const tab = el.querySelector("button[role=tab]");
    expect(tab?.textContent?.trim()).toBe("Notes");
    expect(tab?.className).toContain("sum-notebook-tab--active");
  });

  it("renders contact row in sum_title", () => {
    const record = new SwcRecord("core.user", 1, { name: "Mitchell", email: "m@x.com", phone: "555" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [
          {
            class: "sum_title",
            h1Fields: [{ name: "name", placeholder: "Name" }],
            divs: [
              {
                class: "sum-title-contact-row",
                fields: [
                  { name: "email", string: "Email", widget: "email" },
                  { name: "phone", string: "Work phone" },
                ],
              },
            ],
          },
        ],
        fields: [],
        groups: [],
      },
      record,
      readonly: true,
      hasImageField: true,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    expect(el.querySelector(".sum-title-contact-row")).toBeTruthy();
    expect(el.textContent).toContain("m@x.com");
    expect(el.textContent).toContain("555");
  });

  it("lays out nested groups on a 12-column row with equal spans", () => {
    const record = new SwcRecord("crm.lead", 1, {});
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [],
        fields: [],
        groups: [
          {
            fields: [],
            groups: [
              { string: "A", fields: [{ name: "a", string: "A", type: "char" }] },
              { string: "B", fields: [{ name: "b", string: "B", type: "char" }] },
            ],
          },
        ],
      },
      record,
      readonly: true,
      hasImageField: false,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    const spans = el.querySelectorAll(".sum-form-group-span");
    expect(spans.length).toBe(2);
    expect((spans[0] as HTMLElement).style.getPropertyValue("--sum-group-span")).toBe("6");
    expect((spans[1] as HTMLElement).style.getPropertyValue("--sum-group-span")).toBe("6");
  });

  it("respects col and colspan on group rows", () => {
    const record = new SwcRecord("crm.lead", 1, {});
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [],
        fields: [],
        groups: [
          {
            col: 3,
            fields: [],
            groups: [
              { string: "One", fields: [{ name: "one", type: "char" }] },
              { string: "Two", fields: [{ name: "two", type: "char" }] },
              { string: "Wide", colspan: 2, fields: [{ name: "wide", type: "char" }] },
            ],
          },
        ],
      },
      record,
      readonly: true,
      hasImageField: false,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    const rows = el.querySelectorAll(".sum-form-group-row");
    expect(rows.length).toBe(2);
    const firstRowSpans = rows[0].querySelectorAll(".sum-form-group-span");
    expect(firstRowSpans.length).toBe(2);
    expect((firstRowSpans[0] as HTMLElement).style.getPropertyValue("--sum-group-span")).toBe("4");
    expect((firstRowSpans[1] as HTMLElement).style.getPropertyValue("--sum-group-span")).toBe("4");
    const secondSpan = rows[1].querySelector(".sum-form-group-span") as HTMLElement;
    expect(secondSpan.style.getPropertyValue("--sum-group-span")).toBe("8");
  });

  it("shows hero placeholder when title field is empty", () => {
    const record = new SwcRecord("crm.lead", 1, { name: "" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [
          {
            class: "sum_title",
            h1Fields: [{ name: "name", placeholder: "Opportunity..." }],
          },
        ],
        fields: [],
        groups: [],
      },
      record,
      readonly: false,
      hasImageField: false,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    const input = el.querySelector(".sum-form-hero-input") as HTMLInputElement;
    expect(input).toBeTruthy();
    expect(input.placeholder).toBe("Opportunity...");
    expect(input.value).toBe("");
  });

  it("shows hero placeholder text in readonly mode when empty", () => {
    const record = new SwcRecord("crm.lead", 1, { name: "" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [
          {
            class: "sum_title",
            h1Fields: [{ name: "name", placeholder: "Opportunity..." }],
          },
        ],
        fields: [],
        groups: [],
      },
      record,
      readonly: true,
      hasImageField: false,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    const hero = el.querySelector(".sum-form-hero-input--placeholder");
    expect(hero?.textContent).toBe("Opportunity...");
  });

  it("switches notebook tab panel content", () => {
    const record = new SwcRecord("my.module", 1, { description: "Hello notes" });
    const sheet: SwcArchSheet = {
      divs: [],
      fields: [],
      groups: [],
      notebook: [
        {
          pages: [
            { title: "Description", fields: [{ name: "description", type: "text", placeholder: "Internal notes..." }], groups: [] },
            { title: "Tags", fields: [{ name: "tag_ids", widget: "many2many_tags" }], groups: [] },
          ],
        },
      ],
    };
    const first = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet,
      record,
      readonly: false,
      hasImageField: false,
      activeNotebookPages: { 0: 0 },
      onNotebookTab: () => {},
    }).render();
    expect(first.querySelector(".sum-notebook-tab--active")?.textContent?.trim()).toBe("Description");
    expect(first.querySelector(".sum-field-textarea")).toBeTruthy();
    expect(first.textContent).toContain("Hello notes");

    const second = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet,
      record,
      readonly: false,
      hasImageField: false,
      activeNotebookPages: { 0: 1 },
      onNotebookTab: () => {},
    }).render();
    expect(second.querySelector(".sum-notebook-tab--active")?.textContent?.trim()).toBe("Tags");
    expect(second.querySelector(".sum-field-textarea")).toBeNull();
  });

  it("renders sheet label hint with for attribute", () => {
    const record = new SwcRecord("my.module", 1, { email: "a@x.com" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [],
        fields: [{ name: "email", string: "Email", widget: "email" }],
        groups: [],
        labels: [{ for: "email", string: "Email is rendered above with widget=email" }],
      },
      record,
      readonly: true,
      hasImageField: false,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    const label = el.querySelector("label.sum-form-label--hint");
    expect(label?.getAttribute("for")).toBe("f-email");
    expect(label?.textContent).toContain("widget=email");
  });

  it("renders separator and label on sheet", () => {
    const record = new SwcRecord("my.module", 1, { name: "Demo" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [],
        fields: [],
        groups: [],
        separators: [{ string: "Label example" }],
        labels: [{ for: "email", string: "Email is rendered above" }],
      },
      record,
      readonly: true,
      hasImageField: false,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    expect(el.querySelector(".sum-separator--title")?.textContent).toBe("Label example");
    expect(el.querySelector(".sum-label--notebook, .sum-form-label--hint")?.textContent).toBe("Email is rendered above");
  });
});
