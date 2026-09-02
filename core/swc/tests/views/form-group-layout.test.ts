import { describe, expect, it } from "vitest";
import { gridSpan12, packGroupRows } from "../../src/views/form/form-group-layout.js";
import type { SwcArchGroup } from "../../src/types/workspace.js";

describe("form-group-layout", () => {
  it("gridSpan12 maps colspan within maxCols onto a 12-column grid", () => {
    expect(gridSpan12(2, 1)).toBe(6);
    expect(gridSpan12(3, 2)).toBe(8);
  });

  it("packGroupRows wraps nested groups when colspan exceeds parent col", () => {
    const parent: SwcArchGroup = { col: 2, groups: [], fields: [] };
    const nested: SwcArchGroup[] = [
      { string: "A", colspan: 1, fields: [], groups: [] },
      { string: "B", colspan: 2, fields: [], groups: [] },
    ];
    const rows = packGroupRows(parent, nested);
    expect(rows).toHaveLength(2);
    expect(rows[0]).toHaveLength(1);
    expect(rows[1]).toHaveLength(1);
    expect(rows[0][0].gridSpan).toBe(6);
  });
});
