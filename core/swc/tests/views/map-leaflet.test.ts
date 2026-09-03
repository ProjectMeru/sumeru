import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createElement } from "../harness/dom.js";
import { mountLeafletMap } from "../../src/views/map/map-leaflet.js";

describe("map-leaflet", () => {
  const marker = vi.fn().mockReturnValue({
    addTo: vi.fn().mockReturnThis(),
    bindPopup: vi.fn().mockReturnThis(),
    on: vi.fn().mockReturnThis(),
  });
  const mapRemove = vi.fn();
  const mapSetView = vi.fn();
  const fitBounds = vi.fn();
  const tileLayer = vi.fn().mockReturnValue({ addTo: vi.fn() });

  beforeEach(() => {
    (window as unknown as { L: unknown }).L = {
      map: vi.fn().mockReturnValue({
        remove: mapRemove,
        setView: mapSetView,
        fitBounds,
      }),
      tileLayer,
      marker,
      latLngBounds: vi.fn(),
    };
  });

  afterEach(() => {
    delete (window as unknown as { L?: unknown }).L;
    document.head.innerHTML = "";
  });

  it("mounts markers and returns teardown", async () => {
    const container = createElement("div");
    const onSelect = vi.fn();
    const teardown = await mountLeafletMap(
      container,
      [{ id: 1, lat: 12.97, lng: 77.59, label: "Office" }],
      onSelect,
    );
    expect(marker).toHaveBeenCalled();
    expect(mapSetView).toHaveBeenCalled();
    teardown();
    expect(mapRemove).toHaveBeenCalled();
  });

  it("fitBounds when multiple markers", async () => {
    const container = createElement("div");
    await mountLeafletMap(
      container,
      [
        { id: 1, lat: 10, lng: 20, label: "A" },
        { id: 2, lat: 11, lng: 21, label: "B" },
      ],
      vi.fn(),
    );
    expect(fitBounds).toHaveBeenCalled();
  });

  it("uses cached leaflet when already on window", async () => {
    const container = createElement("div");
    await mountLeafletMap(container, [], vi.fn());
    await mountLeafletMap(container, [], vi.fn());
    expect((window as unknown as { L: { map: ReturnType<typeof vi.fn> } }).L.map).toHaveBeenCalledTimes(2);
  });
});
