import { initPasswordMatchGroups } from "../../login/password-match.js";

function onNotebookKeydown(ev: Event): void {
  if (!(ev instanceof KeyboardEvent)) return;
  if (ev.key !== "ArrowLeft" && ev.key !== "ArrowRight") return;
  const target = ev.target;
  if (!(target instanceof HTMLButtonElement) || target.getAttribute("role") !== "tab") return;
  const tabs = target.parentElement;
  if (!tabs) return;
  const buttons = Array.from(tabs.querySelectorAll<HTMLButtonElement>('button[role="tab"]'));
  const idx = buttons.indexOf(target);
  if (idx < 0) return;
  ev.preventDefault();
  const next = ev.key === "ArrowRight" ? Math.min(idx + 1, buttons.length - 1) : Math.max(idx - 1, 0);
  buttons[next]?.focus();
  buttons[next]?.click();
}

function bindMany2OneDismiss(root: HTMLElement): () => void {
  const onDocClick = (ev: MouseEvent): void => {
    const target = ev.target;
    if (!(target instanceof Node)) return;
    for (const widget of root.querySelectorAll(".sum-field-widget--many2one")) {
      if (widget.contains(target)) continue;
      const list = widget.querySelector(".sum-m2o-suggest");
      list?.remove();
    }
  };
  const onKey = (ev: KeyboardEvent): void => {
    if (ev.key !== "Escape") return;
    for (const list of root.querySelectorAll(".sum-m2o-suggest")) {
      list.remove();
    }
  };
  document.addEventListener("click", onDocClick, true);
  document.addEventListener("keydown", onKey);
  return () => {
    document.removeEventListener("click", onDocClick, true);
    document.removeEventListener("keydown", onKey);
  };
}

interface CropState {
  x: number;
  y: number;
  zoom: number;
}

function normalizeCrop(c: CropState): CropState {
  return {
    x: Math.min(100, Math.max(0, c.x)),
    y: Math.min(100, Math.max(0, c.y)),
    zoom: Math.min(4, Math.max(1, c.zoom)),
  };
}

function applyCropStyle(img: HTMLImageElement, crop: CropState): void {
  const c = normalizeCrop(crop);
  img.style.objectPosition = `${c.x}% ${c.y}%`;
  img.style.transform = `scale(${c.zoom})`;
  img.style.transformOrigin = `${c.x}% ${c.y}%`;
}

function openAvatarCropModal(
  file: File,
  onSave: (dataUrl: string, crop: CropState) => void,
): void {
  const modal = document.createElement("div");
  modal.className = "sum-avatar-crop-modal";
  modal.innerHTML = `
    <div class="sum-avatar-crop-modal-inner">
      <h3 class="sum-avatar-crop-title">Crop image</h3>
      <p class="sum-avatar-crop-hint">Drag to reposition · use zoom slider</p>
      <div class="sum-avatar-crop-stage">
        <div class="sum-avatar-crop-viewport">
          <img class="sum-avatar-crop-img" alt="" />
          <div class="sum-avatar-crop-ring"></div>
        </div>
      </div>
      <label class="sum-avatar-crop-zoom-label">Zoom
        <input type="range" class="sum-avatar-crop-zoom" min="1" max="4" step="0.05" value="1" />
      </label>
      <div class="sum-avatar-crop-modal-actions">
        <button type="button" class="sum-avatar-crop-save">Save</button>
        <button type="button" class="sum-avatar-crop-cancel">Cancel</button>
      </div>
    </div>`;

  const img = modal.querySelector<HTMLImageElement>(".sum-avatar-crop-img")!;
  const zoom = modal.querySelector<HTMLInputElement>(".sum-avatar-crop-zoom")!;
  const stage = modal.querySelector<HTMLDivElement>(".sum-avatar-crop-stage")!;
  let crop: CropState = { x: 50, y: 50, zoom: 1 };
  let dragging = false;

  const close = (): void => modal.remove();

  const reader = new FileReader();
  reader.onload = () => {
    img.src = String(reader.result ?? "");
    applyCropStyle(img, crop);
  };
  reader.readAsDataURL(file);

  stage.addEventListener("pointerdown", (ev) => {
    dragging = true;
    stage.setPointerCapture(ev.pointerId);
  });
  stage.addEventListener("pointermove", (ev) => {
    if (!dragging) return;
    const rect = stage.getBoundingClientRect();
    crop.x = ((ev.clientX - rect.left) / rect.width) * 100;
    crop.y = ((ev.clientY - rect.top) / rect.height) * 100;
    applyCropStyle(img, crop);
  });
  stage.addEventListener("pointerup", () => {
    dragging = false;
  });

  zoom.addEventListener("input", () => {
    crop.zoom = Number(zoom.value);
    applyCropStyle(img, crop);
  });

  modal.querySelector(".sum-avatar-crop-cancel")?.addEventListener("click", close);
  modal.querySelector(".sum-avatar-crop-save")?.addEventListener("click", () => {
    const canvas = document.createElement("canvas");
    const size = 256;
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext("2d");
    if (ctx && img.complete) {
      const c = normalizeCrop(crop);
      const sw = img.naturalWidth / c.zoom;
      const sh = img.naturalHeight / c.zoom;
      const sx = (c.x / 100) * img.naturalWidth - sw / 2;
      const sy = (c.y / 100) * img.naturalHeight - sh / 2;
      ctx.drawImage(img, sx, sy, sw, sh, 0, 0, size, size);
      onSave(canvas.toDataURL("image/png"), c);
    } else {
      onSave(img.src, normalizeCrop(crop));
    }
    close();
  });

  document.body.appendChild(modal);
}

function bindAvatarUpload(root: HTMLElement): () => void {
  const onChange = (ev: Event): void => {
    const input = ev.target;
    if (!(input instanceof HTMLInputElement) || input.type !== "file") return;
    const file = input.files?.[0];
    if (!file || !file.type.startsWith("image/")) return;
    const host =
      input.closest("[data-sum-avatar]") ?? input.closest(".sum-field-widget--image");
    if (!host) return;
    const hidden = host.querySelector<HTMLInputElement>("[data-sum-avatar-value], [data-sum-image-value]");
    openAvatarCropModal(file, (dataUrl) => {
      if (hidden) hidden.value = dataUrl;
      hidden?.dispatchEvent(new Event("input", { bubbles: true }));
      const img = host.querySelector<HTMLImageElement>(".sum-form-avatar-img, .sum-image-thumb-img");
      if (img) {
        img.src = dataUrl;
        img.classList.add("sum-form-avatar-img--visible", "sum-form-avatar-img--cropped");
      }
      const initials = host.querySelector(".sum-form-avatar-initials");
      initials?.remove();
    });
    input.value = "";
  };
  root.addEventListener("change", onChange);
  return () => root.removeEventListener("change", onChange);
}

function bindDateDismiss(root: HTMLElement): () => void {
  const onDocClick = (ev: MouseEvent): void => {
    const target = ev.target;
    if (!(target instanceof Node)) return;
    for (const details of root.querySelectorAll<HTMLDetailsElement>("details.sum-date-field[open]")) {
      if (details.contains(target)) continue;
      details.open = false;
    }
  };
  document.addEventListener("click", onDocClick, true);
  return () => document.removeEventListener("click", onDocClick, true);
}

export function initFormInteractions(root: HTMLElement): () => void {
  const cleanups: Array<() => void> = [];
  for (const tabs of root.querySelectorAll(".sum-notebook-tabs")) {
    tabs.addEventListener("keydown", onNotebookKeydown);
    cleanups.push(() => tabs.removeEventListener("keydown", onNotebookKeydown));
  }
  cleanups.push(bindMany2OneDismiss(root));
  cleanups.push(bindAvatarUpload(root));
  cleanups.push(bindDateDismiss(root));
  initPasswordMatchGroups(root);
  return () => {
    for (const fn of cleanups) fn();
  };
}
