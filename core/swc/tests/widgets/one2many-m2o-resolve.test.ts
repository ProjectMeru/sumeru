import { describe, expect, it, vi } from "vitest";
import { resolveM2oDisplayNames } from "../../src/widgets/one2many-m2o-resolve.js";

describe("resolveM2oDisplayNames", () => {
  it("fills _name fields from searchRead", async () => {
    const rows = [{ partner_id: 1 }];
    const searchRead = vi.fn().mockResolvedValue([{ id: 1, name: "Acme" }]);
    await resolveM2oDisplayNames(
      [{ name: "partner_id", type: "many2one" }],
      rows,
      () => "core.partner",
      searchRead,
    );
    expect(rows[0].partner_id_name).toBe("Acme");
    expect(searchRead).toHaveBeenCalledWith("core.partner", [["id", "in", [1]]], ["id", "name"], 2);
  });

  it("skips non-many2one columns and rows without ids", async () => {
    const searchRead = vi.fn();
    await resolveM2oDisplayNames(
      [{ name: "name", type: "char" }],
      [{ name: "x" }],
      () => "",
      searchRead,
    );
    expect(searchRead).not.toHaveBeenCalled();
  });

  it("continues when searchRead throws", async () => {
    const searchRead = vi.fn().mockRejectedValue(new Error("fail"));
    await expect(
      resolveM2oDisplayNames(
        [{ name: "partner_id", type: "many2one" }],
        [{ partner_id: 2 }],
        () => "core.partner",
        searchRead,
      ),
    ).resolves.toBeUndefined();
  });
});
