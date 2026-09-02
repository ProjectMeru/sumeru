import { describe, expect, it } from "vitest";
import {
  KANBAN_COLUMNS_PER_ROW_DEFAULT,
  kanbanColumnsPerRow,
  kanbanColumnsStyle,
} from "../../src/views/kanban/kanban-layout.js";

describe("kanban-layout", () => {
  it("defaults to 4 columns per row", () => {
    expect(kanbanColumnsPerRow(undefined)).toBe(KANBAN_COLUMNS_PER_ROW_DEFAULT);
    expect(kanbanColumnsPerRow("")).toBe(4);
  });

  it("clamps configured columns per row", () => {
    expect(kanbanColumnsPerRow(6)).toBe(6);
    expect(kanbanColumnsPerRow(0)).toBe(4);
    expect(kanbanColumnsPerRow(99)).toBe(12);
  });

  it("emits css variable style", () => {
    expect(kanbanColumnsStyle(6)).toBe("--sum-kanban-columns:6");
  });
});
