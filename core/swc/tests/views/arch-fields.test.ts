import { describe, expect, it } from "vitest";
import {
  fieldByName,
  graphAxes,
  kanbanCardFields,
  listColumns,
  pivotFields,
  visibleArchFields,
  visibleFields,
} from "../../src/views/shared/arch-fields.js";
import type { SwcViewArch } from "../../src/types/workspace.js";

function arch(partial: Partial<SwcViewArch>): SwcViewArch {
  return {
    type: "list",
    model: "demo.model",
    fields: [],
    ...partial,
  };
}

describe("arch-fields", () => {
  it("visibleArchFields omits invisible fields from a field array", () => {
    expect(
      visibleArchFields([{ name: "a" }, { name: "b", invisible: true }]).map((f) => f.name),
    ).toEqual(["a"]);
  });

  it("visibleFields omits invisible arch fields", () => {
    const fields = visibleFields(
      arch({
        fields: [
          { name: "name" },
          { name: "color", invisible: true },
        ],
      }),
    );
    expect(fields.map((f) => f.name)).toEqual(["name"]);
  });

  it("listColumns matches visible fields", () => {
    const a = arch({ fields: [{ name: "a" }, { name: "b", invisible: true }] });
    expect(listColumns(a).map((f) => f.name)).toEqual(["a"]);
  });

  it("fieldByName finds by name", () => {
    const a = arch({ fields: [{ name: "email" }] });
    expect(fieldByName(a, "email")?.name).toBe("email");
    expect(fieldByName(a, "missing")).toBeUndefined();
  });

  it("kanbanCardFields hides group field and gender", () => {
    const a = arch({
      kanban: { groupField: "state" },
      fields: [
        { name: "state" },
        { name: "gender", invisible: true },
        { name: "name" },
        { name: "email" },
      ],
    });
    expect(kanbanCardFields(a).map((f) => f.name)).toEqual(["name", "email"]);
  });

  it("kanbanCardFields can include invisible fields for color logic", () => {
    const a = arch({
      fields: [
        { name: "color", invisible: true },
        { name: "name" },
      ],
    });
    expect(kanbanCardFields(a, { includeInvisible: true }).map((f) => f.name)).toEqual([
      "color",
      "name",
    ]);
  });

  it("graphAxes reads pivotType row and measure", () => {
    const axes = graphAxes(
      arch({
        graph: { chart: "line" },
        fields: [
          { name: "partner_id", pivotType: "row" },
          { name: "amount", pivotType: "measure" },
        ],
      }),
    );
    expect(axes).toEqual({
      chart: "line",
      groupField: "partner_id",
      measureField: "amount",
    });
  });

  it("graphAxes falls back to defaults", () => {
    expect(graphAxes(arch({ fields: [] }))).toEqual({
      chart: "bar",
      groupField: "create_date",
      measureField: "id",
    });
  });

  it("pivotFields groups row, col, and measure fields", () => {
    const groups = pivotFields(
      arch({
        fields: [
          { name: "partner_id", pivotType: "row" },
          { name: "month", pivotType: "col" },
          { name: "amount", pivotType: "measure" },
        ],
      }),
    );
    expect(groups).toEqual({
      rowFields: ["partner_id"],
      colFields: ["month"],
      measureFields: ["amount"],
    });
  });
});
