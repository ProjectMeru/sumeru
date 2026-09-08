import { describe, expect, it } from "vitest";
import {
  columnsForField,
  displayCellValue,
  formatNumericValue,
  inverseFieldName,
  parseCellValue,
  serverLineValues,
} from "../../src/widgets/one2many-line-utils.js";
import type { SwcArchField } from "../../src/types/workspace.js";
import { configureNumberFormat } from "../../src/i18n/number.js";

describe("one2many-line-utils", () => {
  it("serverLineValues strips id, display names, and empty values", () => {
    expect(
      serverLineValues({
        id: 5,
        name: "Line",
        partner_id: 1,
        partner_id_name: "Acme",
        note: "",
        active: null,
      }),
    ).toEqual({ name: "Line", partner_id: 1 });
  });

  it("formatNumericValue adds thousand separators", () => {
    expect(formatNumericValue(12000)).toBe("12,000");
    expect(formatNumericValue(12000.5)).toBe("12,000.5");
  });

  it("uses configured decimal and grouping separators", () => {
    configureNumberFormat({ decimalPoint: ",", thousandsSep: ".", grouping: "[3,0]" });
    expect(formatNumericValue(1850000.5)).toBe("1.850.000,5");

    configureNumberFormat({ decimalPoint: ".", thousandsSep: ",", grouping: "[3,2,0]" });
    expect(formatNumericValue(123456789)).toBe("12,34,56,789");
  });

  it("displayCellValue formats numbers and prefers _name for relations", () => {
    const col: SwcArchField = { name: "amount", type: "integer" };
    expect(displayCellValue(col, { amount: 1000 })).toBe("1,000");
    const m2o: SwcArchField = { name: "partner_id", type: "many2one" };
    expect(displayCellValue(m2o, { partner_id: 1, partner_id_name: "Acme" })).toBe("Acme");
  });

  it("parseCellValue and helpers coerce line input", () => {
    expect(parseCellValue({ name: "qty", type: "integer" }, "2")).toBe(2);
    expect(parseCellValue({ name: "note", type: "char" }, "")).toBeNull();
    expect(parseCellValue({ name: "done", type: "boolean" }, "1")).toBe(true);
    expect(inverseFieldName("crm.lead")).toBe("lead_id");
    expect(columnsForField({ name: "lines", subview: { fields: [{ name: "name" }] } })).toHaveLength(1);
  });
});
