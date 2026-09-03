import { describe, expect, it, vi } from "vitest";
import { ChatterPanel } from "../../src/views/chatter/ChatterPanel.js";
import { collectionEnv } from "../harness/view.js";

describe("ChatterPanel", () => {
  it("shows empty state for unsaved records", () => {
    const panel = new ChatterPanel({ model: "demo.model", recordId: 0, csrfToken: "tok" }, collectionEnv());
    panel.callSetup();
    const el = panel.render();
    expect(el.classList.contains("sum-chatter--empty")).toBe(true);
    panel.destroy();
  });

  it("loads messages and posts a reply", async () => {
    const env = collectionEnv({
      http: {
        getJSON: vi.fn().mockResolvedValue({
          messages: [{ body: "Hello", author: "Admin", createDate: "2026-01-01" }],
          attachments: [{ id: 1, name: "doc.pdf", url: "/doc.pdf" }],
          enabled: true,
        }),
        postForm: vi.fn().mockResolvedValue(undefined),
        postJSON: vi.fn(),
      },
    });
    const panel = new ChatterPanel({ model: "demo.model", recordId: 5, csrfToken: "tok" }, env);
    panel.callSetup();
    await vi.waitFor(() => {
      expect(panel.render().querySelector(".sum-chatter-message")?.textContent).toContain("Hello");
    });
    const textarea = panel.render().querySelector(".sum-chatter-input") as HTMLTextAreaElement;
    textarea.value = "Reply";
    textarea.dispatchEvent(new Event("input", { bubbles: true }));
    (panel.render().querySelector(".sum-chatter-send") as HTMLButtonElement).click();
    await vi.waitFor(() => expect(env.services.http.postForm).toHaveBeenCalled());
    panel.destroy();
  });

  it("switches to attachments tab", async () => {
    const env = collectionEnv({
      http: {
        getJSON: vi.fn().mockResolvedValue({
          messages: [],
          attachments: [{ id: 2, name: "file.txt", url: "/file.txt" }],
          enabled: true,
        }),
        postForm: vi.fn(),
        postJSON: vi.fn(),
      },
    });
    const panel = new ChatterPanel({ model: "demo.model", recordId: 3, csrfToken: "tok" }, env);
    panel.callSetup();
    await vi.waitFor(() => expect(panel.render().querySelector(".sum-chatter")).toBeTruthy());
    const tabs = panel.render().querySelectorAll(".sum-chatter-tab");
    (tabs[1] as HTMLButtonElement).click();
    expect(panel.render().querySelector(".sum-chatter-attachments")?.textContent).toContain("file.txt");
    panel.destroy();
  });
});
