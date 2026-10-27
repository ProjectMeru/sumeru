import { html, type TemplateResult } from "../../template/html.js";
import type { SwcUserSecurityMeta } from "../../types/workspace.js";

export type UserSecuritySectionMode = "access" | "password";

export function renderUserSecuritySection(
  meta: SwcUserSecurityMeta,
  mode: UserSecuritySectionMode,
  readonly: boolean,
): TemplateResult {
  if (!meta.canEdit || readonly) {
    if (mode === "access") {
      return html`<p class="sum-muted sum-user-security-hint">Group and company assignment requires system administrator.</p>`;
    }
    return html``;
  }

  if (mode === "access") {
    const typeGroups = meta.groups.filter((g) => g.isUserType);
    const otherGroups = meta.groups.filter((g) => !g.isUserType);
    const selectedType = typeGroups.find((g) => g.selected)?.id ?? typeGroups[0]?.id ?? 0;
    const exclusive = new Map<number, typeof otherGroups>();
    const plainGroups = otherGroups.filter((g) => {
      if (!g.exclusive || !g.categoryId) return true;
      const list = exclusive.get(g.categoryId) ?? [];
      list.push(g);
      exclusive.set(g.categoryId, list);
      return false;
    });

    return html`<div class="sum-user-security-root sum-user-security-root--access" data-sum-user-security="access">
      <input type="hidden" name="security_groups_touched" value="1"/>
      ${typeGroups.length > 0
        ? html`<div class="sum-form-group-outer sum-field-region--sheet">
            <div class="sum-form-group-title">User type</div>
            <div class="sum-user-security-radio-group" role="radiogroup">
              ${typeGroups.map(
                (g) => html`<label class="sum-field-radio">
                  <input
                    type="radio"
                    name="security_user_type"
                    value=${String(g.id)}
                    checked=${g.id === selectedType ? "checked" : undefined}
                  />
                  ${g.name}
                </label>`,
              )}
            </div>
          </div>`
        : ""}
      ${[...exclusive.entries()].map(([categoryId, groups]) => {
        const selected = groups.find((g) => g.selected)?.id;
        const title = groups[0]?.categoryName || "Group";
        return html`<div class="sum-form-group-outer sum-field-region--sheet">
          <div class="sum-form-group-title">${title}</div>
          <div class="sum-user-security-radio-group" role="radiogroup">
            ${groups.map(
              (g) => html`<label class="sum-field-radio">
                <input
                  type="radio"
                  name=${"security_exclusive_" + String(categoryId)}
                  value=${String(g.id)}
                  checked=${g.id === selected ? "checked" : undefined}
                />
                ${g.name}
              </label>`,
            )}
          </div>
        </div>`;
      })}
      ${plainGroups.length > 0
        ? html`<div class="sum-form-group-outer sum-field-region--sheet">
            <div class="sum-form-group-title">Groups</div>
            <ul class="sum-user-security-checklist">
              ${plainGroups.map(
                (g) => html`<li>
                  <label class="sum-field-checkbox">
                    <input
                      type="checkbox"
                      name="security_group_ids"
                      value=${String(g.id)}
                      checked=${g.selected ? "checked" : undefined}
                    />
                    ${g.name}
                  </label>
                </li>`,
              )}
            </ul>
          </div>`
        : ""}
      ${meta.companies.length > 0
        ? html`<div class="sum-form-group-outer sum-field-region--sheet">
            <div class="sum-form-group-title">Allowed companies</div>
            <ul class="sum-user-security-checklist">
              ${meta.companies.map(
                (c) => html`<li>
                  <label class="sum-field-checkbox">
                    <input
                      type="checkbox"
                      name="company_ids"
                      value=${String(c.id)}
                      checked=${c.selected ? "checked" : undefined}
                    />
                    ${c.name}
                  </label>
                </li>`,
              )}
            </ul>
          </div>`
        : ""}
    </div>`;
  }

  const required = meta.isNew ? "required" : undefined;
  return html`<div class="sum-user-security-root sum-user-security-root--password" data-sum-user-security="password" data-password-match>
    <div class="sum-form-group-outer sum-field-region--sheet">
      <div class="sum-form-group-title">${meta.passwordLabel}</div>
      <div class="sum-form-group-grid">
        <label class="sum-field-label" for="password_plain">Password</label>
        <input
          id="password_plain"
          name="password_plain"
          type="password"
          autocomplete="new-password"
          class="sum-field-input"
          data-password-primary
          ${required ? html`required=${required}` : ""}
        />
        <label class="sum-field-label" for="password_plain_confirm">Confirm</label>
        <input
          id="password_plain_confirm"
          name="password_plain_confirm"
          type="password"
          autocomplete="new-password"
          class="sum-field-input"
          data-password-confirm
          ${required ? html`required=${required}` : ""}
        />
        <p class="sum-field-hint" data-password-match-hint role="alert" aria-live="polite" hidden></p>
      </div>
    </div>
  </div>`;
}

export type UserSecurityFormFields = Record<string, string | string[]>;

/** Validates initial password fields captured before save (new core.user). */
export function validateNewUserSecurityFields(fields: UserSecurityFormFields | null): string | null {
  const pw = String(fields?.password_plain ?? "").trim();
  const confirm = String(fields?.password_plain_confirm ?? "").trim();
  if (!pw) {
    return "Set an initial password on the Account Security tab before saving.";
  }
  if (pw !== confirm) {
    return "Passwords do not match.";
  }
  return null;
}

/** Collect security fields from the form view root for POST /web/user/security-post. */
export function gatherUserSecurityFields(root: HTMLElement | null): UserSecurityFormFields {
  const out: Record<string, string | string[]> = {};
  if (!root) return out;
  const sections = root.querySelectorAll<HTMLElement>(".sum-user-security-root");
  if (sections.length === 0) return out;

  const add = (name: string, value: string): void => {
    if (name.endsWith("[]")) return;
    if (name === "security_group_ids" || name === "company_ids") {
      const arr = (out[name] as string[] | undefined) ?? [];
      arr.push(value);
      out[name] = arr;
      return;
    }
    out[name] = value;
  };

  for (const section of sections) {
    section.querySelectorAll<HTMLInputElement>("input[name]").forEach((input) => {
      if (input.type === "checkbox" && !input.checked) return;
      if (input.type === "radio" && !input.checked) return;
      if (input.type === "password" && input.value.trim() === "") {
        return;
      }
      if (input.type === "hidden" || input.type === "checkbox" || input.type === "radio" || input.type === "password") {
        add(input.name, input.value);
      }
    });
  }
  return out;
}
