import { describe, expect, it, beforeEach } from "vitest";
import { _t, isRTL, loadTranslations } from "../../src/i18n/translate.js";

describe("translate", () => {
  beforeEach(() => {
    loadTranslations(undefined);
  });

  it("loadTranslations and _t resolve keys", () => {
    loadTranslations({ Save: "Enregistrer" });
    expect(_t("Save")).toBe("Enregistrer");
    expect(_t("Missing")).toBe("Missing");
  });

  it("isRTL reads document direction", () => {
    document.documentElement.dir = "ltr";
    expect(isRTL()).toBe(false);
    document.documentElement.dir = "rtl";
    expect(isRTL()).toBe(true);
    document.documentElement.dir = "ltr";
  });
});
