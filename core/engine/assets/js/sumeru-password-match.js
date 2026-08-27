"use strict";
(() => {
  // src/login/password-match.ts
  var DEFAULT_MESSAGE = "Passwords do not match.";
  function resolveHint(confirm, hint) {
    if (hint) {
      return hint;
    }
    const el = document.createElement("p");
    el.className = "sum-field-hint";
    el.setAttribute("role", "alert");
    el.setAttribute("aria-live", "polite");
    el.hidden = true;
    const field = confirm.closest(".field, .sum-field-widget");
    if (field) {
      field.appendChild(el);
    } else {
      confirm.insertAdjacentElement("afterend", el);
    }
    return el;
  }
  function showMismatch(password, confirm, hint, message) {
    hint.textContent = message;
    hint.hidden = false;
    hint.classList.add("sum-field-hint--error");
    password.classList.add("sum-input-invalid");
    confirm.classList.add("sum-input-invalid");
    confirm.setAttribute("aria-invalid", "true");
  }
  function clearMismatch(password, confirm, hint) {
    hint.textContent = "";
    hint.hidden = true;
    hint.classList.remove("sum-field-hint--error");
    password.classList.remove("sum-input-invalid");
    confirm.classList.remove("sum-input-invalid");
    confirm.removeAttribute("aria-invalid");
  }
  function bindPasswordMatch(options) {
    const { password, confirm, message = DEFAULT_MESSAGE } = options;
    const hint = resolveHint(confirm, options.hint);
    const evaluate = () => {
      if (!confirm.value && !password.value) {
        clearMismatch(password, confirm, hint);
        return true;
      }
      if (password.value !== confirm.value) {
        showMismatch(password, confirm, hint, message);
        return false;
      }
      clearMismatch(password, confirm, hint);
      return true;
    };
    const onInput = () => {
      evaluate();
    };
    password.addEventListener("input", onInput);
    confirm.addEventListener("input", onInput);
    password.addEventListener("blur", onInput);
    confirm.addEventListener("blur", onInput);
    return {
      isValid: evaluate,
      destroy: () => {
        password.removeEventListener("input", onInput);
        confirm.removeEventListener("input", onInput);
        password.removeEventListener("blur", onInput);
        confirm.removeEventListener("blur", onInput);
      }
    };
  }
  function initPasswordMatchGroups(root = document) {
    root.querySelectorAll("[data-password-match]").forEach((group) => {
      if (group.dataset.passwordMatchBound === "1") {
        return;
      }
      const password = group.querySelector("[data-password-primary]");
      const confirm = group.querySelector("[data-password-confirm]");
      if (!password || !confirm) {
        return;
      }
      const hint = group.querySelector("[data-password-match-hint]");
      bindPasswordMatch({ password, confirm, hint });
      group.dataset.passwordMatchBound = "1";
    });
  }

  // src/login/password-match-entry.ts
  window.sumInitPasswordMatch = initPasswordMatchGroups;
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
      initPasswordMatchGroups(document);
    });
  } else {
    initPasswordMatchGroups(document);
  }
})();
//# sourceMappingURL=sumeru-password-match.js.map
