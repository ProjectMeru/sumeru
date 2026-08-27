import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import { forEach } from "../../template/helpers.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_FORM, VIEW_MAP } from "../../constants/routes.js";
import {
  CollectionBarHost,
  mountCollectionBar,
} from "../shared/collection-bar-host.js";

interface MapViewProps {
  payload: SwcWorkspacePayload;
}

const OPEN_STREET_MAP_URL = "https://www.openstreetmap.org/";

function numberField(
  row: Record<string, unknown>,
  name: string,
): number | null {
  const raw = row[name];
  if (raw == null || raw === "") return null;
  const n = Number(raw);
  return Number.isFinite(n) ? n : null;
}

export class MapView extends SwcComponent<MapViewProps> {
  private collectionBar!: CollectionBarHost;

  override setup(): void {
    this.collectionBar = mountCollectionBar(
      this.props.payload,
      VIEW_MAP,
      this.env,
    );
  }

  override onPropsChanged(props: MapViewProps): void {
    this.collectionBar.updateProps({
      payload: props.payload,
      viewType: VIEW_MAP,
    });
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
  }

  private latField(): string {
    return (
      this.props.payload.arch.map?.latitude ||
      this.props.payload.arch.fields.find((f) => /lat/i.test(f.name))?.name ||
      "latitude"
    );
  }

  private lngField(): string {
    return (
      this.props.payload.arch.map?.longitude ||
      this.props.payload.arch.fields.find((f) => /lng|lon/i.test(f.name))
        ?.name ||
      "longitude"
    );
  }

  private openRecord(row: Record<string, unknown>): void {
    const id = Number(row.id ?? 0);
    if (id <= 0) return;
    const payload = this.props.payload;
    this.env.services.action.openRecord({
      actionId: payload.actionId,
      menuId: payload.menuId,
      recordId: id,
      viewType: VIEW_FORM,
    });
  }

  override template() {
    const latName = this.latField();
    const lngName = this.lngField();
    const markers = (this.props.payload.records ?? [])
      .map((row) => {
        const lat = numberField(row, latName);
        const lng = numberField(row, lngName);
        if (lat == null || lng == null) return null;
        return { row, lat, lng };
      })
      .filter(
        (m): m is { row: Record<string, unknown>; lat: number; lng: number } =>
          m != null,
      );

    return html`
      <div class="sum-collection-view sum-map-view">
        ${this.collectionBar.renderOrPatch()}
        <h2>${this.props.payload.arch.title ?? "Map"}</h2>
        <p class="sum-map-hint">${markers.length} located record(s).</p>
        <ul class="sum-map-list">
          ${forEach(
            markers,
            (marker) => Number(marker.row.id ?? 0),
            (marker) =>
              html`<li class="sum-map-item">
                <button
                  type="button"
                  class="sum-map-name"
                  @click=${() => this.openRecord(marker.row)}
                >
                  ${String(
                    marker.row.name ?? marker.row.display_name ?? marker.row.id,
                  )}
                </button>
                <a
                  class="sum-map-link"
                  href=${`${OPEN_STREET_MAP_URL}?mlat=${marker.lat}&mlon=${marker.lng}#map=16/${marker.lat}/${marker.lng}`}
                  target="_blank"
                  rel="noopener"
                  >${marker.lat.toFixed(4)}, ${marker.lng.toFixed(4)}</a
                >
              </li>`,
          )}
        </ul>
      </div>
    `;
  }
}
