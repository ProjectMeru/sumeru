import { describe, expect, it } from "vitest";
import { formatFieldValue } from "../../src/views/shared/field-display.js";
import type { SwcArchField } from "../../src/types/workspace.js";

function field(partial: Partial<SwcArchField> & Pick<SwcArchField, "name">): SwcArchField {
  return { ...partial, name: partial.name };
}

describe("formatFieldValue", () => {
  it("uses many2one display name when present", () => {
    const row = { partner_id: 3, partner_id_name: "Acme Corp" };
    expect(formatFieldValue(row, field({ name: "partner_id", type: "many2one" }))).toBe("Acme Corp");
  });

  it("formats boolean and selection values", () => {
    expect(formatFieldValue({ active: true }, field({ name: "active", type: "boolean" }))).toBe("Yes");
    expect(formatFieldValue({ active: false }, field({ name: "active", type: "boolean" }))).toBe("No");
    expect(
      formatFieldValue(
        { state: "done" },
        field({
          name: "state",
          type: "selection",
          selection: [
            ["draft", "Draft"],
            ["done", "Done"],
          ],
        }),
      ),
    ).toBe("Done");
  });
});
