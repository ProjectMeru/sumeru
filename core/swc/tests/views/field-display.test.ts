import { describe, expect, it } from "vitest";
import { formatFieldValue, recordDisplayLabel } from "../../src/views/shared/field-display.js";

describe("recordDisplayLabel", () => {
  it("prefers name then display_name", () => {
    expect(recordDisplayLabel({ name: "Acme" })).toBe("Acme");
    expect(recordDisplayLabel({ display_name: "Display" })).toBe("Display");
  });

  it("uses fallback or id hash", () => {
    expect(recordDisplayLabel({}, 5)).toBe("5");
    expect(recordDisplayLabel({ id: 9 })).toBe("#9");
  });
});

describe("formatFieldValue", () => {
  it("formats boolean fields", () => {
    expect(formatFieldValue({ active: true }, { name: "active", type: "boolean" })).toBe("Yes");
    expect(formatFieldValue({ active: false }, { name: "active", type: "boolean" })).toBe("No");
  });
});
