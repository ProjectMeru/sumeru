export const IMAGE_PLACEHOLDER_URLS = {
  generic: "/static/img/image_placeholder.jpg",
  male: "/static/img/male_person.jpg",
  female: "/static/img/female_person.jpg",
  building: "/static/img/office_building_placeholder.jpg",
} as const;

export interface ImagePlaceholderContext {
  model?: string;
  gender?: unknown;
  isCompany?: unknown;
}

export function isUploadedImageSrc(src: string): boolean {
  const v = src.trim();
  if (!v) return false;
  return (
    v.startsWith("data:") ||
    v.startsWith("http://") ||
    v.startsWith("https://") ||
    v.startsWith("/")
  );
}

function isTruthyFlag(value: unknown): boolean {
  if (value === true || value === 1) return true;
  if (typeof value === "string") {
    const v = value.trim().toLowerCase();
    return v === "true" || v === "1" || v === "yes" || v === "t";
  }
  return false;
}

function normalizeGender(value: unknown): string {
  if (value == null || value === false || value === "") return "";
  return String(value).trim().toLowerCase();
}

export function resolvePlaceholderSrc(ctx: ImagePlaceholderContext): string {
  const model = (ctx.model ?? "").trim().toLowerCase();
  if (model === "core.company" || isTruthyFlag(ctx.isCompany)) {
    return IMAGE_PLACEHOLDER_URLS.building;
  }
  switch (normalizeGender(ctx.gender)) {
    case "female":
      return IMAGE_PLACEHOLDER_URLS.female;
    case "male":
      return IMAGE_PLACEHOLDER_URLS.male;
    default:
      return IMAGE_PLACEHOLDER_URLS.generic;
  }
}

/** Uploaded image wins; otherwise contextual default placeholder. */
export function resolveImageDisplaySrc(raw: unknown, ctx: ImagePlaceholderContext = {}): string {
  if (typeof raw === "string" && isUploadedImageSrc(raw)) {
    return raw.trim();
  }
  return resolvePlaceholderSrc(ctx);
}

export function isPlaceholderDisplaySrc(src: string, ctx: ImagePlaceholderContext = {}): boolean {
  return src === resolvePlaceholderSrc(ctx);
}
