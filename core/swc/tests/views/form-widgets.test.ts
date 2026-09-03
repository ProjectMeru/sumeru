import { describe, expect, it, vi } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import { registerDefaultWidgets, renderField } from "../../src/widgets/registry.js";
import type { SwcArchField } from "../../src/types/workspace.js";
import type { SwcEnv } from "../../src/runtime/env.js";

registerDefaultWidgets();

const env = {
  bootstrap: {} as never,
  services: {
    rpc: {
      searchRead: vi.fn().mockResolvedValue([]),
    },
  },
} as unknown as SwcEnv;

function field(overrides: Partial<SwcArchField>): SwcArchField {
  return { name: "x", ...overrides };
}

describe("form field widgets", () => {
  it("default char field emits sum-field-widget shell", () => {
    const record = new SwcRecord("m", 1, { name: "Acme" });
    const el = renderField(env, field({ name: "name", string: "Name", type: "char" }), record, false);
    expect(el.classList.contains("sum-field-widget")).toBe(true);
    expect(el.querySelector(".sum-field-input")).toBeTruthy();
  });

  it("phone widget uses tel input", () => {
    const record = new SwcRecord("m", 1, { phone: "555" });
    const el = renderField(env, field({ name: "phone", widget: "phone" }), record, false);
    expect(el.querySelector('input[type="tel"]')).toBeTruthy();
  });

  it("date field uses inline native date input", () => {
    const record = new SwcRecord("m", 1, { date_start: "2026-01-01" });
    const el = renderField(env, field({ name: "date_start", type: "date", string: "Start Date" }), record, false);
    expect(el.querySelector(".sum-date-field-inline")).toBeTruthy();
    expect(el.querySelector(".sum-date-input")?.getAttribute("type")?.trim()).toBe("date");
    expect(el.querySelector(".sum-date-action-btn")).toBeTruthy();
  });

  it("date field Today and Clear buttons update record", () => {
    const record = new SwcRecord("m", 1, { date_start: "2026-01-01" });
    const el = renderField(env, field({ name: "date_start", type: "date" }), record, false);
    const buttons = el.querySelectorAll(".sum-date-action-btn");
    (buttons[1] as HTMLButtonElement).click();
    expect(record.get("date_start")).toBeNull();
    (buttons[0] as HTMLButtonElement).click();
    expect(record.get("date_start")).toBeTruthy();
  });

  it("datetime field uses datetime-local input", () => {
    const record = new SwcRecord("m", 1, { when: "2026-01-01T10:30:00Z" });
    const el = renderField(env, field({ name: "when", type: "datetime" }), record, false);
    expect(el.querySelector('.sum-date-input[type="datetime-local"]')).toBeTruthy();
  });

  it("many2many_tags renders tag container", () => {
    const record = new SwcRecord("m", 1, { tag_ids: [1] });
    const el = renderField(
      env,
      field({ name: "tag_ids", widget: "many2many_tags", relation: "my.module.tag" }),
      record,
      true,
    );
    expect(el.querySelector(".sum-field-tags, .sum-multi-select-tags")).toBeTruthy();
  });

  it("radio widget renders radio group", () => {
    const record = new SwcRecord("m", 1, { active: true });
    const el = renderField(env, field({ name: "active", type: "boolean", widget: "radio" }), record, false);
    expect(el.querySelector(".sum-field-radio-group")).toBeTruthy();
  });

  it("boolean checkbox widget toggles value", () => {
    const record = new SwcRecord("m", 1, { active: true });
    const el = renderField(env, field({ name: "active", type: "boolean" }), record, false);
    const input = el.querySelector('input[type="checkbox"]') as HTMLInputElement;
    expect(input.checked).toBe(true);
    input.checked = false;
    input.dispatchEvent(new Event("change", { bubbles: true }));
    expect(record.get("active")).toBe(false);
  });

  it("boolean readonly shows Yes/No", () => {
    const record = new SwcRecord("m", 1, { active: false });
    const el = renderField(env, field({ name: "active", type: "boolean" }), record, true);
    expect(el.textContent).toContain("No");
  });

  it("image widget shows placeholder when empty", () => {
    const record = new SwcRecord("core.partner", 1, { avatar: "" });
    const el = renderField(env, field({ name: "avatar", widget: "image" }), record, false);
    expect(el.querySelector(".sum-image-thumb-img")?.getAttribute("src")).toContain(
      "/static/img/image_placeholder.jpg",
    );
    expect(el.querySelector(".sum-image-upload-hint")).toBeTruthy();
  });

  it("image widget shows thumbnail when value set", () => {
    const record = new SwcRecord("m", 1, { avatar: "/img/x.png" });
    const el = renderField(env, field({ name: "avatar", widget: "image" }), record, true);
    expect(el.querySelector(".sum-image-thumb-img")?.getAttribute("src")).toBe("/img/x.png");
  });

  it("many2one widget renders search input", async () => {
    const rpc = env.services.rpc as { searchRead: ReturnType<typeof vi.fn> };
    rpc.searchRead.mockResolvedValue([{ id: 1, name: "Acme" }]);
    const record = new SwcRecord("m", 1, { partner_id: null });
    const el = renderField(
      env,
      field({ name: "partner_id", type: "many2one", relation: "res.partner", string: "Partner" }),
      record,
      false,
    );
    expect(el.querySelector(".sum-m2o-wrap")).toBeTruthy();
    const input = el.querySelector(".sum-field-input") as HTMLInputElement;
    input.value = "Ac";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await vi.waitFor(() => expect(rpc.searchRead).toHaveBeenCalled());
  });

  it("extra field widgets render readonly and editable modes", () => {
    const monetary = new SwcRecord("m", 1, { amount: "10" });
    expect(
      renderField(env, field({ name: "amount", widget: "monetary", options: { currency_symbol: "$" } }), monetary, true)
        .textContent,
    ).toContain("$ 10");

    const htmlRec = new SwcRecord("m", 1, { body: "<b>Hi</b>" });
    expect(renderField(env, field({ name: "body", widget: "html" }), htmlRec, true).textContent).toContain("Hi");

    const binary = new SwcRecord("m", 1, { doc: "file.pdf", doc_name: "Report.pdf" });
    expect(renderField(env, field({ name: "doc", widget: "binary" }), binary, true).querySelector("a")).toBeTruthy();

    const color = new SwcRecord("m", 1, { idx: 3 });
    expect(renderField(env, field({ name: "idx", widget: "color" }), color, true).querySelector(".sum-color-swatch")).toBeTruthy();

    const urlRec = new SwcRecord("m", 1, { website: "https://example.com" });
    const urlEl = renderField(env, field({ name: "website", widget: "url" }), urlRec, true);
    expect(urlEl.querySelector('a[href="https://example.com"]')).toBeTruthy();

    const progress = new SwcRecord("m", 1, { pct: 40 });
    expect(renderField(env, field({ name: "pct", widget: "progress" }), progress, true).textContent).toContain("40%");

    expect(renderField(env, field({ name: "seq", widget: "handle" }), new SwcRecord("m", 1, {}), false).querySelector(".sum-handle-grip")).toBeTruthy();
  });

  it("boolean_toggle widget shows On/Off label", () => {
    const record = new SwcRecord("m", 1, { active: true });
    const el = renderField(env, field({ name: "active", widget: "boolean_toggle", string: "Active" }), record, false);
    expect(el.textContent).toContain("On");
    expect(el.querySelector(".sum-field-toggle")).toBeTruthy();
  });

  it("selection field renders static options", async () => {
    const record = new SwcRecord("m", 1, { state: "draft" });
    const el = renderField(
      env,
      field({
        name: "state",
        type: "selection",
        selection: [
          ["draft", "Draft"],
          ["done", "Done"],
        ],
      }),
      record,
      false,
    );
    await vi.waitFor(() => expect(el.querySelector("select")).toBeTruthy());
    expect(el.querySelector("option")).toBeTruthy();
  });

  it("boolean radio sets false when No selected", () => {
    const record = new SwcRecord("m", 1, { active: true });
    const el = renderField(env, field({ name: "active", type: "boolean", widget: "radio" }), record, false);
    const no = el.querySelector('input[value="0"]') as HTMLInputElement;
    no.checked = true;
    no.dispatchEvent(new Event("change", { bubbles: true }));
    expect(record.get("active")).toBe(false);
  });

  it("textarea and integer fields accept edits", () => {
    const textRec = new SwcRecord("m", 1, { note: "hi" });
    const textEl = renderField(env, field({ name: "note", type: "text" }), textRec, false);
    const ta = textEl.querySelector("textarea") as HTMLTextAreaElement;
    ta.value = "updated";
    ta.dispatchEvent(new Event("input", { bubbles: true }));
    expect(textRec.get("note")).toBe("updated");

    const intRec = new SwcRecord("m", 1, { qty: 1 });
    const intEl = renderField(env, field({ name: "qty", type: "integer" }), intRec, false);
    const input = intEl.querySelector("input") as HTMLInputElement;
    input.value = "5";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    expect(intRec.get("qty")).toBe(5);
  });

  it("phone field fires change handler", () => {
    const record = new SwcRecord("m", 1, { phone: "555" });
    const el = renderField(env, field({ name: "phone", widget: "phone" }), record, false);
    const input = el.querySelector('input[type="tel"]') as HTMLInputElement;
    input.value = "999";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    expect(record.get("phone")).toBe("999");
  });

  it("boolean toggle change updates record", () => {
    const record = new SwcRecord("m", 1, { active: true });
    const el = renderField(env, field({ name: "active", widget: "boolean_toggle" }), record, false);
    const input = el.querySelector('input[type="checkbox"]') as HTMLInputElement;
    input.checked = false;
    input.dispatchEvent(new Event("change", { bubbles: true }));
    expect(record.get("active")).toBe(false);
  });
});
