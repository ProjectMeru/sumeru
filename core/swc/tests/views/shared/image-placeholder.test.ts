import { describe, expect, it } from "vitest";
import {
  IMAGE_PLACEHOLDER_URLS,
  isUploadedImageSrc,
  resolveImageDisplaySrc,
  resolvePlaceholderSrc,
} from "../../../src/views/shared/image-placeholder.js";

describe("image-placeholder", () => {
  it("detects uploaded image sources", () => {
    expect(isUploadedImageSrc("/static/x.jpg")).toBe(true);
    expect(isUploadedImageSrc("data:image/png;base64,abc")).toBe(true);
    expect(isUploadedImageSrc("")).toBe(false);
    expect(isUploadedImageSrc("not-a-url")).toBe(false);
  });

  it("prefers uploaded image over placeholders", () => {
    expect(resolveImageDisplaySrc("/uploads/a.png", { gender: "male" })).toBe("/uploads/a.png");
  });

  it("uses building placeholder for company records", () => {
    expect(resolvePlaceholderSrc({ model: "core.company" })).toBe(IMAGE_PLACEHOLDER_URLS.building);
    expect(resolvePlaceholderSrc({ isCompany: true })).toBe(IMAGE_PLACEHOLDER_URLS.building);
  });

  it("uses gender placeholders for people", () => {
    expect(resolvePlaceholderSrc({ gender: "female" })).toBe(IMAGE_PLACEHOLDER_URLS.female);
    expect(resolvePlaceholderSrc({ gender: "male" })).toBe(IMAGE_PLACEHOLDER_URLS.male);
  });

  it("falls back to generic placeholder", () => {
    expect(resolvePlaceholderSrc({})).toBe(IMAGE_PLACEHOLDER_URLS.generic);
    expect(resolvePlaceholderSrc({ gender: "other" })).toBe(IMAGE_PLACEHOLDER_URLS.generic);
  });
});
