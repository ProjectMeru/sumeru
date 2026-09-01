import { describe, expect, it, afterEach } from "vitest";
import { DialogService } from "../../src/services/dialog.js";

describe("DialogService", () => {
  const dialog = new DialogService();

  afterEach(() => {
    dialog.close(false);
    document.body.innerHTML = "";
  });

  it("confirm resolves true when OK clicked", async () => {
    const promise = dialog.confirm("Title", "Body");
    const ok = document.querySelector(".sum-dialog-btn--primary") as HTMLButtonElement;
    ok.click();
    await expect(promise).resolves.toBe(true);
  });

  it("alert resolves after OK", async () => {
    const promise = dialog.alert("Title", "Body");
    document.querySelector(".sum-dialog-btn--primary")?.dispatchEvent(new MouseEvent("click"));
    await expect(promise).resolves.toBeUndefined();
  });

  it("openHost renders custom content", async () => {
    const host = document.createElement("div");
    host.textContent = "Custom";
    const promise = dialog.openHost("Host", host);
    expect(document.querySelector(".sum-dialog-body--host")?.textContent).toContain("Custom");
    document.querySelector(".sum-dialog-btn")?.dispatchEvent(new MouseEvent("click"));
    await expect(promise).resolves.toBe(false);
  });

  it("Escape dismisses with false", async () => {
    const promise = dialog.open({ title: "T", body: "B" });
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    await expect(promise).resolves.toBe(false);
  });
});
