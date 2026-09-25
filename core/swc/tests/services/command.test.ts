import { describe, expect, it, vi } from "vitest";
import { CommandService } from "../../src/services/command.js";

describe("CommandService", () => {
  it("registers and runs commands", () => {
    const cmd = new CommandService();
    const run = vi.fn();
    cmd.register({ id: "test", label: "Test", run });
    expect(cmd.run("test")).toBe(true);
    expect(run).toHaveBeenCalled();
  });

  it("filters by when()", () => {
    const cmd = new CommandService();
    cmd.register({ id: "a", label: "A", run: () => {}, when: () => false });
    cmd.register({ id: "b", label: "B", run: () => {}, when: () => true });
    expect(cmd.list().map((c) => c.id)).toEqual(["b"]);
  });
});
