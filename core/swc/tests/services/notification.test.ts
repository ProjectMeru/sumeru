import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { NotificationService } from "../../src/services/notification.js";

describe("NotificationService", () => {
  let stack: HTMLElement;
  let svc: NotificationService;

  beforeEach(() => {
    stack = document.createElement("div");
    stack.id = "sum-toast-stack";
    stack.className = "sum-toast-stack";
    stack.setAttribute("aria-live", "polite");
    document.body.appendChild(stack);
    svc = new NotificationService(stack);
  });

  afterEach(() => {
    vi.useRealTimers();
    stack.remove();
  });

  it("renders title, body, close, and kind class", () => {
    svc.show({ kind: "info", title: "Hello", body: "World" });
    const toast = stack.querySelector(".sum-toast") as HTMLElement;
    expect(toast.classList.contains("sum-toast--info")).toBe(true);
    expect(toast.querySelector(".sum-toast-title")?.textContent).toBe("Hello");
    expect(toast.querySelector(".sum-toast-body")?.textContent).toBe("World");
    expect(toast.querySelector(".sum-toast-close")).toBeTruthy();
  });

  it("renders optional details", () => {
    svc.show({ kind: "error", title: "Failed", body: "Write failed", details: "column x" });
    expect(stack.querySelector(".sum-toast-details")?.textContent).toBe("column x");
  });

  it("success, error, and warning helpers set kind", () => {
    svc.success("Saved", "ok");
    svc.error("Broken", "no");
    svc.warning("Careful", "maybe");
    const kinds = [...stack.querySelectorAll(".sum-toast")].map((el) => el.className);
    expect(kinds[0]).toContain("sum-toast--success");
    expect(kinds[1]).toContain("sum-toast--error");
    expect(kinds[2]).toContain("sum-toast--warning");
  });

  it("close starts exit animation and removes the node", () => {
    svc.show({ kind: "info", title: "T", body: "B" }, 60_000);
    const toast = stack.querySelector(".sum-toast") as HTMLElement;
    (toast.querySelector(".sum-toast-close") as HTMLButtonElement).click();
    expect(toast.classList.contains("sum-toast-out")).toBe(true);
    toast.dispatchEvent(new Event("animationend"));
    expect(stack.children.length).toBe(0);
  });

  it("caps the stack at 5 and dismisses the oldest first", () => {
    for (let i = 1; i <= 6; i++) {
      svc.show({ kind: "info", title: String(i), body: "n" }, 60_000);
    }
    const titles = [...stack.querySelectorAll(".sum-toast-title")].map((el) => el.textContent);
    expect(titles).toContain("1");
    const first = [...stack.querySelectorAll(".sum-toast")].find(
      (el) => el.querySelector(".sum-toast-title")?.textContent === "1",
    );
    expect(first?.classList.contains("sum-toast-out")).toBe(true);
    const live = [...stack.querySelectorAll(".sum-toast")].filter(
      (el) => !el.classList.contains("sum-toast-out"),
    );
    expect(live).toHaveLength(5);
  });

  it("pauses auto-dismiss while hovered", () => {
    vi.useFakeTimers();
    svc.show({ kind: "info", title: "Hover", body: "wait" }, 6000);
    const toast = stack.querySelector(".sum-toast") as HTMLElement;
    vi.advanceTimersByTime(3000);
    toast.dispatchEvent(new Event("mouseenter"));
    vi.advanceTimersByTime(10_000);
    expect(toast.classList.contains("sum-toast-out")).toBe(false);
    toast.dispatchEvent(new Event("mouseleave"));
    vi.advanceTimersByTime(3000);
    expect(toast.classList.contains("sum-toast-out")).toBe(true);
  });

  it("bootstrap replays server messages and creates stack when missing", () => {
    stack.remove();
    const auto = new NotificationService();
    auto.bootstrap([{ kind: "info", title: "Boot", body: "msg" }]);
    expect(document.getElementById("sum-toast-stack")?.querySelector(".sum-toast-title")?.textContent).toBe("Boot");
    document.getElementById("sum-toast-stack")?.remove();
  });
});
