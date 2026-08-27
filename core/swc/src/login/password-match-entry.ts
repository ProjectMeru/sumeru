import { initPasswordMatchGroups } from "./password-match.js";

declare global {
  interface Window {
    sumInitPasswordMatch?: typeof initPasswordMatchGroups;
  }
}

window.sumInitPasswordMatch = initPasswordMatchGroups;

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => {
    initPasswordMatchGroups(document);
  });
} else {
  initPasswordMatchGroups(document);
}
