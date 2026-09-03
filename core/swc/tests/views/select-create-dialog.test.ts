import { afterEach, describe, expect, it, vi } from "vitest";
import { SelectCreateDialog } from "../../src/views/dialogs/select-create-dialog.js";
import { collectionEnv } from "../harness/view.js";

describe("SelectCreateDialog", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("searches and selects a row", async () => {
    const onSelect = vi.fn();
    const env = collectionEnv({
      rpc: {
        searchRead: vi.fn().mockResolvedValue([{ id: 9, name: "Partner" }]),
        write: vi.fn(),
        create: vi.fn(),
        unlink: vi.fn(),
        read: vi.fn(),
        call: vi.fn(),
      },
    });
    SelectCreateDialog.open(env, { comodel: "core.partner", onSelect });
    await vi.waitFor(() => {
      expect(document.querySelector(".sum-select-create-item")?.textContent).toContain("Partner");
    });
    (document.querySelector(".sum-select-create-item") as HTMLButtonElement).click();
    expect(onSelect).toHaveBeenCalledWith({ id: 9, name: "Partner" });
    expect(document.querySelector(".sum-modal-host")).toBeNull();
  });

  it("creates a new record from footer action", async () => {
    const onSelect = vi.fn();
    const env = collectionEnv({
      rpc: {
        searchRead: vi.fn().mockResolvedValue([]),
        create: vi.fn().mockResolvedValue(12),
        write: vi.fn(),
        unlink: vi.fn(),
        read: vi.fn(),
        call: vi.fn(),
      },
    });
    SelectCreateDialog.open(env, { comodel: "core.partner", onSelect });
    await vi.waitFor(() => expect(document.querySelector(".sum-select-create")).toBeTruthy());
    const input = document.querySelector(".sum-field-input") as HTMLInputElement;
    input.value = "NewCo";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    (document.querySelector(".sum-modal-footer .sum-btn") as HTMLButtonElement).click();
    await vi.waitFor(() => expect(onSelect).toHaveBeenCalledWith({ id: 12, name: "NewCo" }));
  });

  it("closes on backdrop click", async () => {
    const onCancel = vi.fn();
    SelectCreateDialog.open(collectionEnv(), { comodel: "core.partner", onSelect: vi.fn(), onCancel });
    await vi.waitFor(() => expect(document.querySelector(".sum-modal-backdrop")).toBeTruthy());
    (document.querySelector(".sum-modal-backdrop") as HTMLElement).click();
    expect(onCancel).toHaveBeenCalled();
    expect(document.querySelector(".sum-modal-host")).toBeNull();
  });
});
