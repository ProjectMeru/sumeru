import { describe, expect, it } from "vitest";
import { parseArchDate, resolveArchDateField } from "../../src/views/shared/arch-date.js";
import type { SwcViewArch } from "../../src/types/workspace.js";

function arch(partial: Partial<SwcViewArch> = {}): SwcViewArch {
  return {
    type: "calendar",
    model: "demo.model",
    fields: [{ name: "date_deadline", type: "date" }],
    ...partial,
  };
}

describe("arch-date", () => {
  it("parseArchDate returns null for empty or invalid input", () => {
    expect(parseArchDate("")).toBeNull();
    expect(parseArchDate("not-a-date")).toBeNull();
  });

  it("parseArchDate parses ISO date strings", () => {
    const date = parseArchDate("2024-06-15");
    expect(date).toBeInstanceOf(Date);
    expect(date!.getFullYear()).toBe(2024);
  });

  it("resolveArchDateField prefers arch keys in order then first date field", () => {
    expect(resolveArchDateField(arch({ calendar: { dateStart: "start_on" } }), ["calendar"], "x")).toBe("start_on");
    expect(resolveArchDateField(arch(), ["calendar"], "date_deadline")).toBe("date_deadline");
    expect(
      resolveArchDateField(
        arch({ cohort: { dateStart: "cohort_date" }, gantt: { dateStart: "gantt_date" } }),
        ["cohort", "calendar", "gantt"],
        "create_date",
      ),
    ).toBe("cohort_date");
  });
});
