import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchField } from "../../types/workspace.js";

export function isKanbanImageField(field: SwcArchField): boolean {
  const name = field.name.toLowerCase();
  return (
    name === "image" ||
    name.startsWith("image_") ||
    field.widget === "image" ||
    field.widget === "circle"
  );
}

export function isKanbanImageCircle(field: SwcArchField): boolean {
  return field.widget === "circle" || field.options?.shape === "circle";
}

export function isPriorityField(field: SwcArchField): boolean {
  return field.name === "priority" || field.widget === "priority";
}

function displayValue(row: Record<string, unknown>, field: SwcArchField): string {
  const raw = row[`${field.name}_name`] ?? row[field.name];
  if (raw == null || raw === false) return "";
  return String(raw);
}

export function isKanbanCardRotting(row: Record<string, unknown>, thresholdDays = 7): boolean {
  if (thresholdDays <= 0) {
    thresholdDays = 7;
  }
  const raw = row.date_last_stage_update;
  if (raw == null || raw === false || raw === "") {
    return false;
  }
  const text = String(raw).trim();
  const updated = Date.parse(text.length >= 10 ? text.slice(0, 10) : text);
  if (Number.isNaN(updated)) {
    return false;
  }
  const ageMs = Date.now() - updated;
  return ageMs >= thresholdDays * 24 * 60 * 60 * 1000;
}

function renderActivityIndicator(row: Record<string, unknown>): TemplateResult | null {
  const deadline = row.activity_deadline;
  const summary = row.activity_summary;
  if ((deadline == null || deadline === "") && (summary == null || summary === "")) {
    return null;
  }
  const label = summary ? String(summary) : "Activity";
  const when = deadline ? String(deadline) : "";
  return html`<div class="sum-kanban-card-activity" title=${when}>📅 ${label}${when ? html` · ${when}` : ""}</div>`;
}

function imageSrc(row: Record<string, unknown>, field: SwcArchField): string {
  const raw = row[field.name];
  if (typeof raw !== "string" || !raw.trim()) return "";
  const v = raw.trim();
  if (v.startsWith("data:") || v.startsWith("http://") || v.startsWith("https://") || v.startsWith("/")) {
    return v;
  }
  return "";
}

function initials(row: Record<string, unknown>, fields: SwcArchField[]): string {
  const nameField = fields.find((f) => f.name === "name") ?? fields.find((f) => !isKanbanImageField(f));
  const text = nameField ? displayValue(row, nameField) : "";
  const parts = text.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 1).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

function titleField(fields: SwcArchField[]): SwcArchField | undefined {
  return (
    fields.find((f) => f.name === "name") ??
    fields.find((f) => f.name === "display_name") ??
    fields.find((f) => !isKanbanImageField(f) && !isPriorityField(f))
  );
}

function renderPriority(row: Record<string, unknown>, field: SwcArchField): TemplateResult | null {
  const level = Number(row[field.name] ?? 0);
  if (!level) return null;
  const stars = [1, 2, 3].map(
    (n) =>
      html`<span class="sum-kanban-priority-star${n <= level ? " sum-kanban-priority-star--on" : ""}">★</span>`,
  );
  return html`<div class="sum-kanban-priority">${stars}</div>`;
}

function renderMedia(row: Record<string, unknown>, imageField: SwcArchField, fields: SwcArchField[]): HTMLElement | null {
  const src = imageSrc(row, imageField);
  const label = displayValue(row, titleField(fields) ?? imageField);
  if (!src && !label) return null;

  const media = document.createElement("div");
  media.className = `sum-kanban-card-media${isKanbanImageCircle(imageField) ? " sum-kanban-card-media--circle" : " sum-kanban-card-media--square"}`;

  if (src) {
    const img = document.createElement("img");
    img.className = "sum-kanban-card-media-img";
    img.src = src;
    img.alt = "";
    media.appendChild(img);
  } else {
    const initialsEl = document.createElement("span");
    initialsEl.className = "sum-kanban-card-media-initials";
    initialsEl.textContent = initials(row, fields);
    media.appendChild(initialsEl);
  }
  return media;
}

export function renderKanbanCardInner(row: Record<string, unknown>, fields: SwcArchField[]): TemplateResult {
  const imageField = fields.find(isKanbanImageField);
  const priorityField = fields.find(isPriorityField);
  const title = titleField(fields);
  const subs = fields.filter(
    (f) => f !== imageField && f !== title && f !== priorityField && !isKanbanImageField(f) && !isPriorityField(f),
  );

  const media = imageField ? renderMedia(row, imageField, fields) : null;

  const titleEl = title ? html`<div class="sum-kanban-card-title">${displayValue(row, title)}</div>` : null;
  const subEls = subs
    .map((f) => displayValue(row, f))
    .filter(Boolean)
    .map((text) => html`<div class="sum-kanban-card-sub">${text}</div>`);
  const priorityEl = priorityField ? renderPriority(row, priorityField) : null;
  const activityEl = renderActivityIndicator(row);

  if (media) {
    return html`${media}<div class="sum-kanban-card-body">${titleEl}${subEls}${priorityEl}${activityEl}</div>`;
  }

  return html`${titleEl}${subEls}${priorityEl}${activityEl}`;
}
