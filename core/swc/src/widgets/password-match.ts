export interface PasswordMatchBinding {
  isValid: () => boolean;
  destroy: () => void;
}

export interface PasswordMatchOptions {
  password: HTMLInputElement;
  confirm: HTMLInputElement;
  hint?: HTMLElement | null;
  message?: string;
}

const DEFAULT_MESSAGE = "Passwords do not match.";

function resolveHint(
  confirm: HTMLInputElement,
  hint?: HTMLElement | null,
): HTMLElement {
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

function showMismatch(
  password: HTMLInputElement,
  confirm: HTMLInputElement,
  hint: HTMLElement,
  message: string,
): void {
  hint.textContent = message;
  hint.hidden = false;
  hint.classList.add("sum-field-hint--error");
  password.classList.add("sum-input-invalid");
  confirm.classList.add("sum-input-invalid");
  confirm.setAttribute("aria-invalid", "true");
}

function clearMismatch(
  password: HTMLInputElement,
  confirm: HTMLInputElement,
  hint: HTMLElement,
): void {
  hint.textContent = "";
  hint.hidden = true;
  hint.classList.remove("sum-field-hint--error");
  password.classList.remove("sum-input-invalid");
  confirm.classList.remove("sum-input-invalid");
  confirm.removeAttribute("aria-invalid");
}

/** Bind live password / confirm validation on input. */
export function bindPasswordMatch(options: PasswordMatchOptions): PasswordMatchBinding {
  const { password, confirm, message = DEFAULT_MESSAGE } = options;
  const hint = resolveHint(confirm, options.hint);

  const evaluate = (): boolean => {
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

  const onInput = (): void => {
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
    },
  };
}

/** Returns false when any bound group has mismatched non-empty passwords. */
export function validatePasswordMatchGroups(root: ParentNode = document): boolean {
  let ok = true;
  root.querySelectorAll<HTMLElement>("[data-password-match]").forEach((group) => {
    const password = group.querySelector<HTMLInputElement>("[data-password-primary]");
    const confirm = group.querySelector<HTMLInputElement>("[data-password-confirm]");
    if (!password || !confirm) {
      return;
    }
    if (!password.value && !confirm.value) {
      return;
    }
    if (password.value !== confirm.value) {
      const hint = group.querySelector<HTMLElement>("[data-password-match-hint]");
      if (hint) {
        showMismatch(password, confirm, hint, DEFAULT_MESSAGE);
      }
      ok = false;
    }
  });
  return ok;
}

/** Wire all `[data-password-match]` groups under root. */
export function initPasswordMatchGroups(root: ParentNode = document): void {
  root.querySelectorAll<HTMLElement>("[data-password-match]").forEach((group) => {
    if (group.dataset.passwordMatchBound === "1") {
      return;
    }
    const password = group.querySelector<HTMLInputElement>("[data-password-primary]");
    const confirm = group.querySelector<HTMLInputElement>("[data-password-confirm]");
    if (!password || !confirm) {
      return;
    }
    const hint = group.querySelector<HTMLElement>("[data-password-match-hint]");
    bindPasswordMatch({ password, confirm, hint });
    group.dataset.passwordMatchBound = "1";
  });
}
