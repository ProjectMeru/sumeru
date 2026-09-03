import { describe, expect, it, vi } from "vitest";
import { KanbanView } from "../../src/views/kanban/KanbanView.js";
import { collectionEnv, viewPayload } from "../harness/view.js";

describe("KanbanView", () => {
  it("renders flat kanban cards when no columns configured", () => {
    const view = new KanbanView(
      {
        payload: viewPayload(
          { type: "kanban", fields: [{ name: "name" }] },
          [{ id: 1, name: "Card A" }],
        ),
      },
      collectionEnv(),
    );
    view.callSetup();
    const el = view.render();
    expect(el.querySelector(".sum-kanban-card")).toBeTruthy();
    expect(el.textContent).toContain("Card A");
    view.destroy();
  });

  it("renders grouped stages with quick create", async () => {
    const env = collectionEnv();
    const create = env.services.rpc.create as ReturnType<typeof vi.fn>;
    const view = new KanbanView(
      {
        payload: viewPayload({
          type: "kanban",
          fields: [{ name: "name" }],
          kanban: {
            groupField: "stage_id",
            quickCreate: true,
            columns: [{ value: 1, label: "New", color: 2, records: [] }],
          },
        }),
      },
      env,
    );
    view.callSetup();
    const el = view.render();
    const input = el.querySelector(".sum-kanban-quick-input") as HTMLInputElement;
    input.value = "Fresh lead";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    el.querySelector(".sum-kanban-quick-create")?.dispatchEvent(
      new Event("submit", { bubbles: true, cancelable: true }),
    );
    await vi.waitFor(() => expect(create).toHaveBeenCalled());
    view.destroy();
  });

  it("moves card on drop and writes group field", async () => {
    const env = collectionEnv();
    const write = env.services.rpc.write as ReturnType<typeof vi.fn>;
    const view = new KanbanView(
      {
        payload: viewPayload({
          type: "kanban",
          fields: [{ name: "name" }, { name: "color", invisible: true }],
          kanban: {
            groupField: "stage_id",
            draggable: true,
            columns: [
              { value: 1, label: "Todo", color: 0, records: [{ id: 7, name: "Move me" }] },
              { value: 2, label: "Done", color: 1, records: [] },
            ],
          },
        }),
      },
      env,
    );
    view.callSetup();
    const el = view.render();
    const stage = el.querySelectorAll(".sum-kanban-stage-section")[1] as HTMLElement;
    const drop = new Event("drop", { bubbles: true, cancelable: true }) as Event & {
      dataTransfer?: { getData: (key: string) => string };
      preventDefault: () => void;
    };
    drop.dataTransfer = { getData: () => "7" };
    drop.preventDefault = vi.fn();
    stage.dispatchEvent(drop);
    await vi.waitFor(() => expect(write).toHaveBeenCalledWith("demo.model", [7], { stage_id: 2 }));
    view.destroy();
  });

  it("opens color picker and sets card color", async () => {
    const env = collectionEnv();
    const write = env.services.rpc.write as ReturnType<typeof vi.fn>;
    const view = new KanbanView(
      {
        payload: viewPayload(
          {
            type: "kanban",
            fields: [{ name: "name" }, { name: "color", invisible: true }],
          },
          [{ id: 4, name: "Tint", color: false }],
        ),
      },
      env,
    );
    view.callSetup();
    const el = view.render();
    (el.querySelector(".sum-kanban-card-color-btn") as HTMLButtonElement).click();
    const swatch = view.render().querySelector(".sum-kanban-color-swatch") as HTMLButtonElement;
    swatch.click();
    await vi.waitFor(() => expect(write).toHaveBeenCalledWith("demo.model", [4], { color: 0 }));
    view.destroy();
  });
});
