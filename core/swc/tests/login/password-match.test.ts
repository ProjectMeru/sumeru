import { describe, it, expect, afterEach } from "vitest";
import { bindPasswordMatch } from "../../src/login/password-match.js";

describe("bindPasswordMatch", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("shows a hint when passwords differ", () => {
    document.body.innerHTML = `
      <input id="pw" type="password" />
      <input id="pw2" type="password" />
      <p id="hint" hidden></p>
    `;
    const password = document.getElementById("pw") as HTMLInputElement;
    const confirm = document.getElementById("pw2") as HTMLInputElement;
    const hint = document.getElementById("hint") as HTMLElement;
    const binding = bindPasswordMatch({ password, confirm, hint });

    password.value = "secret123";
    confirm.value = "different";
    confirm.dispatchEvent(new Event("input"));

    expect(binding.isValid()).toBe(false);
    expect(hint.hidden).toBe(false);
    expect(hint.textContent).toBe("Passwords do not match.");
    expect(confirm.classList.contains("sum-input-invalid")).toBe(true);
  });

  it("clears the hint when passwords match", () => {
    document.body.innerHTML = `
      <input id="pw" type="password" />
      <input id="pw2" type="password" />
      <p id="hint" hidden></p>
    `;
    const password = document.getElementById("pw") as HTMLInputElement;
    const confirm = document.getElementById("pw2") as HTMLInputElement;
    const hint = document.getElementById("hint") as HTMLElement;
    bindPasswordMatch({ password, confirm, hint });

    password.value = "secret123";
    confirm.value = "secret123";
    confirm.dispatchEvent(new Event("input"));

    expect(hint.hidden).toBe(true);
    expect(confirm.classList.contains("sum-input-invalid")).toBe(false);
  });
});
