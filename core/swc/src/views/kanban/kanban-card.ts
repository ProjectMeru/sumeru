import { html, type TemplateResult } from "../../template/html.js";
import type { SwcArchField } from "../../types/workspace.js";
import { formatFieldValue } from "../shared/field-display.js";
import {
  isPlaceholderDisplaySrc,
  isUploadedImageSrc,
  resolveImageDisplaySrc,
  type ImagePlaceholderContext,
} from "../shared/image-placeholder.js";
import { isKanbanColorField } from "./kanban-color.js";

export function isKanbanImageField(field: SwcArchField): boolean {
  const name = field.name.toLowerCase();
  return (
    name === "image" ||
    name.startsWith("image_") ||
    field.widget === "image" ||
    field.widget === "circle"
  );
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

function imageContext(row: Record<string, unknown>, modelName?: string): ImagePlaceholderContext {
  return {
    model: modelName,
    gender: row.gender,
    isCompany: row.is_company,
  };
}

function titleField(fields: SwcArchField[]): SwcArchField | undefined {
  return (
    fields.find((f) => f.name === "name") ??
    fields.find((f) => f.name === "display_name") ??
    fields.find((f) => !isKanbanImageField(f) && !isPriorityField(f) && !isKanbanColorField(f))
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

function renderMedia(
  row: Record<string, unknown>,
  imageField: SwcArchField,
  modelName?: string,
): HTMLElement {
  const ctx = imageContext(row, modelName);
  const src = resolveImageDisplaySrc(row[imageField.name], ctx);
  const uploaded = typeof row[imageField.name] === "string" && isUploadedImageSrc(String(row[imageField.name]));

  const media = document.createElement("div");
  media.className = "sum-kanban-card-media sum-kanban-card-media--rect";

  const img = document.createElement("img");
  img.className = "sum-kanban-card-media-img";
  img.src = src;
  img.alt = "";
  if (!uploaded || isPlaceholderDisplaySrc(src, ctx)) {
    img.setAttribute("data-sum-image-placeholder", "1");
  }
  media.appendChild(img);
  return media;
}

function renderLabeledField(row: Record<string, unknown>, field: SwcArchField): TemplateResult | null {
  const value = formatFieldValue(row, field);
  if (!value) return null;
  const label = field.string ?? field.name;
  return html`<div class="sum-kanban-card-field">
    <span class="sum-kanban-card-field-label">${label}:</span>
    <span class="sum-kanban-card-field-value">${value}</span>
  </div>`;
}

export function renderKanbanCardInner(
  row: Record<string, unknown>,
  fields: SwcArchField[],
  modelName?: string,
): TemplateResult {
  const imageField = fields.find(isKanbanImageField);
  const priorityField = fields.find(isPriorityField);
  const title = titleField(fields);
  const subs = fields.filter(
    (f) =>
      f !== imageField &&
      f !== title &&
      f !== priorityField &&
      !isKanbanImageField(f) &&
      !isPriorityField(f) &&
      !isKanbanColorField(f) &&
      f.name !== "color" &&
      f.name !== "gender",
  );

  const media = imageField ? renderMedia(row, imageField, modelName) : null;
  const titleEl = title ? html`<div class="sum-kanban-card-title">${displayValue(row, title)}</div>` : null;
  const subEls = subs.map((f) => renderLabeledField(row, f)).filter(Boolean);
  const priorityEl = priorityField ? renderPriority(row, priorityField) : null;
  const activityEl = renderActivityIndicator(row);

  return html`<div class="sum-kanban-card-inner">
    ${media ?? ""}
    <div class="sum-kanban-card-body">${titleEl}${subEls}${priorityEl}${activityEl}</div>
  </div>`;
}
