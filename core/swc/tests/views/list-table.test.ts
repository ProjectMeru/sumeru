import { describe, expect, it, vi } from "vitest";
import { renderArchListTable } from "../../src/views/shared/list-table.js";

describe("renderArchListTable", () => {
  it("renders columns and row values", () => {
    const root = renderArchListTable({
      columns: [
        { name: "name", string: "Name" },
        { name: "email" },
      ],
      rows: [{ id: 1, name: "Acme", email: "a@acme.test" }],
      onRowClick: () => undefined,
    }).render();

    expect(root.querySelector(".sum-list-table")).toBeTruthy();
    expect(root.querySelectorAll(".sum-list-th").length).toBe(2);
    expect(root.textContent).toContain("Acme");
    expect(root.textContent).toContain("a@acme.test");
  });

  it("invokes onRowClick when row is clicked", () => {
    const onRowClick = vi.fn();
    const row = { id: 3, name: "Beta" };
    const root = renderArchListTable({
      columns: [{ name: "name" }],
      rows: [row],
      onRowClick,
    }).render();

    root.querySelector(".sum-list-row")?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    expect(onRowClick).toHaveBeenCalledWith(row);
  });

  it("applies firstCellStyle on the first column", () => {
    const root = renderArchListTable({
      columns: [{ name: "name" }, { name: "email" }],
      rows: [{ id: 1, name: "Acme", email: "x@test" }],
      onRowClick: () => undefined,
      firstCellStyle: () => "padding-left: 2rem",
    }).render();

    const cells = root.querySelectorAll(".sum-list-td");
    expect((cells[0] as HTMLElement).style.paddingLeft).toBe("2rem");
    expect((cells[1] as HTMLElement).style.paddingLeft).toBe("");
  });
});
