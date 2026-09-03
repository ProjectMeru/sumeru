import { describe, expect, it, vi } from "vitest";
import { registerDefaultWidgets } from "../../src/widgets/registry.js";
import { FormView } from "../../src/views/form/FormView.js";
import { collectionEnv, viewPayload } from "../harness/view.js";
import type { SwcWorkspacePayload } from "../../src/types/workspace.js";

registerDefaultWidgets();

function formPayload(recordId: number, formEdit: boolean): SwcWorkspacePayload {
  return {
    ...viewPayload({
      type: "form",
      fields: [{ name: "name", type: "char", required: true }],
      sheet: {
        divs: [],
        fields: [{ name: "name", type: "char", string: "Name" }],
        groups: [],
      },
    }),
    recordId,
    formEdit,
    record: { id: recordId, name: "Acme" },
  };
}

describe("FormView", () => {
  it("renders readonly toolbar for existing records", () => {
    const view = new FormView({ payload: formPayload(5, false) }, collectionEnv());
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-form-view--readonly")).toBeTruthy();
    expect(el.textContent).toContain("Edit");
    view.destroy();
  });

  it("enters edit mode from toolbar", () => {
    const view = new FormView({ payload: formPayload(5, false) }, collectionEnv());
    view.callSetup();
    const buttons = view.render().querySelectorAll(".sum-header-btn");
    const editBtn = [...buttons].find((b) => b.textContent?.trim() === "Edit") as HTMLButtonElement;
    editBtn.click();
    expect(view.render().textContent).toContain("Save");
    view.destroy();
  });

  it("saves new record and opens it", async () => {
    const env = collectionEnv();
    const openRecord = env.services.action.openRecord as ReturnType<typeof vi.fn>;
    (env.services.rpc.create as ReturnType<typeof vi.fn>).mockResolvedValue(42);
    const view = new FormView({ payload: formPayload(0, true) }, env);
    view.callSetup();
    const saveBtn = [...view.render().querySelectorAll(".sum-header-btn")].find(
      (b) => b.textContent?.trim() === "Save",
    ) as HTMLButtonElement;
    saveBtn.click();
    await vi.waitFor(() => expect(openRecord).toHaveBeenCalled());
    view.destroy();
  });

  it("delete confirms and navigates back to list", async () => {
    const env = collectionEnv();
    const navigate = env.services.action.navigate as ReturnType<typeof vi.fn>;
    (env.services.rpc.unlink as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    const view = new FormView({ payload: formPayload(8, false) }, env);
    view.callSetup();
    const deleteBtn = [...view.render().querySelectorAll(".sum-header-btn")].find(
      (b) => b.textContent?.trim() === "Delete",
    ) as HTMLButtonElement;
    deleteBtn.click();
    await vi.waitFor(() => expect(navigate).toHaveBeenCalled());
    view.destroy();
  });
});
