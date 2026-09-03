import { describe, expect, it } from "vitest";
import { isKanbanImageField, isPriorityField, renderKanbanCardInner } from "../../src/views/kanban/kanban-card.js";
import {
  isKanbanColorField,
  kanbanStripeClass,
  resolveCardColor,
} from "../../src/views/kanban/kanban-color.js";

describe("kanban-card", () => {
  it("detects image and priority fields", () => {
    expect(isKanbanImageField({ name: "image", widget: "circle" })).toBe(true);
    expect(isPriorityField({ name: "priority" })).toBe(true);
    expect(isKanbanColorField({ name: "color" })).toBe(true);
  });

  it("renders title and labeled field rows", () => {
    const el = renderKanbanCardInner(
      { id: 1, name: "Acme", email: "a@example.com" },
      [
        { name: "name", string: "Name" },
        { name: "email", string: "Email" },
      ],
    ).render();
    expect(el.querySelector(".sum-kanban-card-title")?.textContent).toBe("Acme");
    const field = el.querySelector(".sum-kanban-card-field");
    expect(field?.querySelector(".sum-kanban-card-field-label")?.textContent).toBe("Email:");
    expect(field?.querySelector(".sum-kanban-card-field-value")?.textContent).toBe("a@example.com");
  });

  it("renders placeholder media when image field present but empty", () => {
    const el = renderKanbanCardInner(
      { id: 1, name: "Acme", image: "", gender: "male" },
      [{ name: "image" }, { name: "name" }],
      "core.user",
    ).render();
    const img = el.querySelector(".sum-kanban-card-media-img");
    expect(img).not.toBeNull();
    expect(img?.getAttribute("src")).toContain("/static/img/male_person.jpg");
    expect(img?.getAttribute("data-sum-image-placeholder")).toBe("1");
  });

  it("renders uploaded image in media block", () => {
    const withImage = renderKanbanCardInner(
      { id: 1, name: "Acme", image: "https://example.com/x.png" },
      [{ name: "image" }, { name: "name" }],
    ).render();
    expect(withImage.querySelector(".sum-kanban-card-media")).not.toBeNull();
    expect(withImage.querySelector(".sum-kanban-card-media-img")?.getAttribute("src")).toBe(
      "https://example.com/x.png",
    );
  });

  it("excludes color field from card body", () => {
    const el = renderKanbanCardInner(
      { id: 1, name: "Acme", color: 3 },
      [
        { name: "color" },
        { name: "name", string: "Name" },
      ],
    ).render();
    expect(el.textContent).not.toContain("Color");
    expect(el.querySelector(".sum-kanban-card-field")).toBeNull();
  });

  it("renders priority stars when priority field present", () => {
    const el = renderKanbanCardInner(
      { id: 1, name: "Acme", priority: 3 },
      [{ name: "priority" }, { name: "name" }],
    ).render();
    expect(el.querySelector(".sum-kanban-priority")).not.toBeNull();
  });

  it("wraps content in inner and body containers", () => {
    const el = renderKanbanCardInner({ id: 1, name: "Acme" }, [{ name: "name" }]).render();
    const inner = el.classList.contains("sum-kanban-card-inner") ? el : el.querySelector(".sum-kanban-card-inner");
    expect(inner).not.toBeNull();
    expect((inner ?? el).querySelector(".sum-kanban-card-body")).not.toBeNull();
  });
});

describe("kanban-color", () => {
  it("prefers record color over stage color", () => {
    expect(resolveCardColor({ color: 5 }, 2)).toBe(5);
  });

  it("falls back to stage color when record color unset", () => {
    expect(resolveCardColor({}, 2)).toBe(2);
    expect(resolveCardColor({ color: false }, 4)).toBe(4);
  });

  it("returns null when no colors set", () => {
    expect(resolveCardColor({}, undefined)).toBeNull();
  });

  it("maps stripe class from color index", () => {
    expect(kanbanStripeClass(3)).toBe("sum-kanban-card-stripe sum-kanban-card-stripe--color-3");
    expect(kanbanStripeClass(null)).toContain("stripe--none");
  });
});
