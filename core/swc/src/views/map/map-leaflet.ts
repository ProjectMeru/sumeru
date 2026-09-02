export interface MapMarker {
  id: number;
  lat: number;
  lng: number;
  label: string;
}

const LEAFLET_CSS = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.css";
const LEAFLET_JS = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.js";

let leafletPromise: Promise<typeof window & { L?: any }> | null = null;

function loadLeaflet(): Promise<any> {
  if (typeof window === "undefined") {
    return Promise.reject(new Error("no window"));
  }
  if ((window as any).L) {
    return Promise.resolve((window as any).L);
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
      script.onload = () => resolve((window as any).L);
      script.onerror = () => reject(new Error("leaflet load failed"));
      document.head.appendChild(script);
    });
  }
  return leafletPromise.then(() => (window as any).L);
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

  const layer: any[] = [];
  for (const marker of markers) {
    const m = L.marker([marker.lat, marker.lng]).addTo(map);
    m.bindPopup(marker.label);
    m.on("click", () => onSelect(marker.id));
    layer.push(m);
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
