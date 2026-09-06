export interface MapMarker {
  id: number;
  lat: number;
  lng: number;
  label: string;
}

/** Minimal Leaflet surface used by mountLeafletMap (loaded from CDN). */
interface LeafletMarker {
  bindPopup(html: string): LeafletMarker;
  on(event: string, fn: () => void): void;
}

interface LeafletMap {
  setView(latlng: [number, number], zoom: number): void;
  fitBounds(bounds: unknown, opts?: { padding?: [number, number] }): void;
  remove(): void;
}

interface LeafletNS {
  map(el: HTMLElement, opts?: { scrollWheelZoom?: boolean }): LeafletMap;
  tileLayer(
    url: string,
    opts?: { attribution?: string; maxZoom?: number },
  ): { addTo(map: LeafletMap): void };
  marker(latlng: [number, number]): LeafletMarker & { addTo(map: LeafletMap): LeafletMarker };
  latLngBounds(points: [number, number][]): unknown;
}

type WindowWithLeaflet = Window & { L?: LeafletNS };

const LEAFLET_CSS = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.css";
const LEAFLET_JS = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.js";

let leafletPromise: Promise<LeafletNS> | null = null;

function leafletFromWindow(): LeafletNS | undefined {
  return (window as WindowWithLeaflet).L;
}

function loadLeaflet(): Promise<LeafletNS> {
  if (typeof window === "undefined") {
    return Promise.reject(new Error("no window"));
  }
  const existing = leafletFromWindow();
  if (existing) {
    return Promise.resolve(existing);
  }
  if (!leafletPromise) {
    leafletPromise = new Promise((resolve, reject) => {
      if (!document.querySelector(`link[href="${LEAFLET_CSS}"]`)) {
        const link = document.createElement("link");
        link.rel = "stylesheet";
        link.href = LEAFLET_CSS;
        document.head.appendChild(link);
      }
      const script = document.createElement("script");
      script.src = LEAFLET_JS;
      script.async = true;
      script.onload = () => {
        const L = leafletFromWindow();
        if (!L) {
          reject(new Error("leaflet missing after load"));
          return;
        }
        resolve(L);
      };
      script.onerror = () => reject(new Error("leaflet load failed"));
      document.head.appendChild(script);
    });
  }
  return leafletPromise;
}

export async function mountLeafletMap(
  container: HTMLElement,
  markers: MapMarker[],
  onSelect: (id: number) => void,
): Promise<() => void> {
  const L = await loadLeaflet();
  const map = L.map(container, { scrollWheelZoom: true });
  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    attribution: "&copy; OpenStreetMap",
    maxZoom: 19,
  }).addTo(map);

  for (const marker of markers) {
    const m = L.marker([marker.lat, marker.lng]).addTo(map);
    m.bindPopup(marker.label);
    m.on("click", () => onSelect(marker.id));
  }

  if (markers.length === 1) {
    map.setView([markers[0].lat, markers[0].lng], 13);
  } else if (markers.length > 1) {
    const bounds = L.latLngBounds(markers.map((m) => [m.lat, m.lng]));
    map.fitBounds(bounds, { padding: [24, 24] });
  } else {
    map.setView([20, 0], 2);
  }

  return () => {
    map.remove();
  };
}
