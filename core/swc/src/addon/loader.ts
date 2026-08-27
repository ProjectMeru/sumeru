import { registry } from "../runtime/registry.js";

/** Loads addon swc_entry modules registered in manifest static paths. */
export class AddonLoader {
  static async loadEntries(urls: string[]): Promise<void> {
    for (const url of urls) {
      try {
        await import(/* @vite-ignore */ url);
      } catch (err) {
        console.warn("SWC addon entry failed:", url, err);
      }
    }
  }

  static registerFromGlobal(): void {
    const entries = (window as unknown as { __SWC_ADDON_ENTRIES__?: string[] }).__SWC_ADDON_ENTRIES__;
    if (entries?.length) {
      void AddonLoader.loadEntries(entries);
    }
  }
}

export { registry };
