import { describe, expect, it } from "vitest";
import { isKanbanImageField, isPriorityField, renderKanbanCardInner } from "../../src/views/kanban/kanban-card.js";

describe("kanban-card", () => {
  it("detects image and priority fields", () => {
    expect(isKanbanImageField({ name: "image", widget: "circle" })).toBe(true);
    expect(isPriorityField({ name: "priority" })).toBe(true);
  });

  it("renders title and subtitle classes", () => {
    const el = renderKanbanCardInner(
      { id: 1, name: "Acme", email: "a@example.com" },
      [
        { name: "name" },
        { name: "email" },
      ],
    ).render();
    expect(el.querySelector(".sum-kanban-card-title")?.textContent).toBe("Acme");
    expect(el.querySelector(".sum-kanban-card-sub")?.textContent).toBe("a@example.com");
  });

  it("renders image media block when image field present", () => {
    const el = renderKanbanCardInner(
      { id: 1, name: "Acme", image: "https://example.com/x.png" },
      [
        { name: "image", widget: "circle" },
        { name: "name" },
      ],
    ).render();
    expect(el.querySelector(".sum-kanban-card-media")).not.toBeNull();
    expect(el.querySelector(".sum-kanban-card-media-img")?.getAttribute("src")).toBe(
      "https://example.com/x.png",
    );
  });

  it("renders priority stars when priority field present", () => {
    const el = renderKanbanCardInner(
      { id: 1, name: "Acme", priority: 3 },
      [{ name: "priority" }, { name: "name" }],
    ).render();
    expect(el.querySelector(".sum-kanban-priority")).not.toBeNull();
  });
});
