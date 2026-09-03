import { describe, expect, it, vi } from "vitest";
import { SwcRecord } from "../../src/model/record.js";
import { registerDefaultWidgets, renderField } from "../../src/widgets/registry.js";
import { renderFormToolbar, renderFormToolbarPrimary } from "../../src/views/form/form-chrome.js";
import { viewPayload } from "../harness/view.js";
import type { SwcEnv } from "../../src/runtime/env.js";

registerDefaultWidgets();

const env = { bootstrap: {} as never, services: {} as never } as SwcEnv;

describe("form-chrome", () => {
  it("readonly toolbar includes New, Edit, Duplicate, Delete", () => {
    const record = new SwcRecord("demo.model", 3, { name: "Acme" });
    const buttons = renderFormToolbarPrimary({
      payload: { ...viewPayload({ type: "form" }), recordId: 3 },
      readonly: true,
      busy: false,
      record,
      fields: [],
      onStartEdit: vi.fn(),
      onCancelEdit: vi.fn(),
      onSave: vi.fn(),
      onDelete: vi.fn(),
      onDuplicate: vi.fn(),
      onObjectButton: vi.fn(),
      renderField: (field, rec, ro) => renderField(env, field, rec, ro),
    });
    expect(buttons.map((b) => b.textContent?.trim())).toEqual(
      expect.arrayContaining(["New", "Edit", "Duplicate", "Delete"]),
    );
  });

  it("dialog readonly toolbar omits New and Delete", () => {
    const record = new SwcRecord("demo.model", 3, { name: "Acme" });
    const buttons = renderFormToolbarPrimary({
      payload: { ...viewPayload({ type: "form" }), recordId: 3 },
      inDialog: true,
      readonly: true,
      busy: false,
      record,
      fields: [],
      onStartEdit: vi.fn(),
      onCancelEdit: vi.fn(),
      onSave: vi.fn(),
      onDelete: vi.fn(),
      onDuplicate: vi.fn(),
      onObjectButton: vi.fn(),
      renderField: (field, rec, ro) => renderField(env, field, rec, ro),
    });
    expect(buttons.map((b) => b.textContent?.trim())).toEqual(["Edit"]);
  });

  it("editing toolbar shows Save and Cancel", () => {
    const record = new SwcRecord("demo.model", 0, { name: "" });
    const buttons = renderFormToolbarPrimary({
      payload: { ...viewPayload({ type: "form" }), recordId: 3 },
      readonly: false,
      busy: false,
      record,
      fields: [],
      onStartEdit: vi.fn(),
      onCancelEdit: vi.fn(),
      onSave: vi.fn(),
      onDelete: vi.fn(),
      onDuplicate: vi.fn(),
      onObjectButton: vi.fn(),
      renderField: (field, rec, ro) => renderField(env, field, rec, ro),
    });
    expect(buttons.map((b) => b.textContent?.trim())).toEqual(["Save", "Cancel"]);
  });

  it("includes object header buttons when visible", () => {
    const record = new SwcRecord("demo.model", 3, { name: "Acme" });
    const onObjectButton = vi.fn();
    const payload = viewPayload({ type: "form" });
    payload.recordId = 3;
    payload.arch.header = {
      buttons: [{ name: "action_confirm", string: "Confirm", type: "object" }],
    };
    const buttons = renderFormToolbarPrimary({
      payload,
      readonly: true,
      busy: false,
      record,
      fields: [],
      onStartEdit: vi.fn(),
      onCancelEdit: vi.fn(),
      onSave: vi.fn(),
      onDelete: vi.fn(),
      onDuplicate: vi.fn(),
      onObjectButton,
      renderField: (field, rec, ro) => renderField(env, field, rec, ro),
    });
    const confirm = buttons.find((b) => b.textContent?.trim() === "Confirm") as HTMLButtonElement;
    confirm.click();
    expect(onObjectButton).toHaveBeenCalled();
  });

  it("renderFormToolbar includes header status fields", () => {
    const record = new SwcRecord("demo.model", 1, { state: "done" });
    const el = renderFormToolbar({
      payload: viewPayload({
        type: "form",
        header: { fields: [{ name: "state", widget: "statusbar" }] },
      }),
      readonly: true,
      busy: false,
      record,
      fields: [{ name: "state", widget: "statusbar" }],
      onStartEdit: vi.fn(),
      onCancelEdit: vi.fn(),
      onSave: vi.fn(),
      onDelete: vi.fn(),
      onDuplicate: vi.fn(),
      onObjectButton: vi.fn(),
      renderField: (field, rec, ro) => renderField(env, field, rec, ro),
    }).render();
    expect(el.querySelector(".sum-statusbar-status")).toBeTruthy();
  });
});
