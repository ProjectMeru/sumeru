import { SwcComponent } from "../../runtime/component.js";
import { html } from "../../template/html.js";
import type { SwcWorkspacePayload } from "../../types/workspace.js";
import { VIEW_FORM, VIEW_MAP } from "../../constants/routes.js";
import { onMount, useTemplateRef } from "../../runtime/hooks.js";
import { onWillUnmount } from "../../runtime/lifecycle.js";
import {
  CollectionBarHost,
  mountCollectionBar,
} from "../shared/collection-bar-host.js";
import { mountLeafletMap, type MapMarker } from "./map-leaflet.js";

interface MapViewProps {
  payload: SwcWorkspacePayload;
}

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
  private mapContainerRef!: { current: Element | null };
  private teardownMap: (() => void) | null = null;

  override setup(): void {
    this.mapContainerRef = useTemplateRef("map-canvas");
    this.collectionBar = mountCollectionBar(
      this.props.payload,
      VIEW_MAP,
      this.env,
    );
    onMount(() => void this.renderMap());
    onWillUnmount(() => {
      this.teardownMap?.();
      this.teardownMap = null;
    });
  }

  override onPropsChanged(props: MapViewProps): void {
    this.collectionBar.updateProps({
      payload: props.payload,
      viewType: VIEW_MAP,
    });
    void this.renderMap();
  }

  override onWillUnmount(): void {
    this.collectionBar.destroy();
    this.teardownMap?.();
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

  private markers(): MapMarker[] {
    const latName = this.latField();
    const lngName = this.lngField();
    return (this.props.payload.records ?? [])
      .map((row) => {
        const lat = numberField(row, latName);
        const lng = numberField(row, lngName);
        if (lat == null || lng == null) return null;
        const id = Number(row.id ?? 0);
        return {
          id,
          lat,
          lng,
          label: String(row.name ?? row.display_name ?? id),
        };
      })
      .filter((m): m is MapMarker => m != null && m.id > 0);
  }

  private openRecord(recordId: number): void {
    const payload = this.props.payload;
    this.env.services.action.openRecord({
      actionId: payload.actionId,
      menuId: payload.menuId,
      recordId,
      viewType: VIEW_FORM,
    });
  }

  private async renderMap(): Promise<void> {
    const el = this.mapContainerRef.current;
    if (!(el instanceof HTMLElement)) return;
    this.teardownMap?.();
    this.teardownMap = null;
    el.innerHTML = "";
    const markers = this.markers();
    try {
      this.teardownMap = await mountLeafletMap(el, markers, (id) =>
        this.openRecord(id),
      );
    } catch {
      el.textContent = `${markers.length} located record(s). Map tiles unavailable.`;
    }
  }

  override template() {
    const count = this.markers().length;
    return html`
      <div class="sum-collection-view sum-map-view">
        ${this.collectionBar.renderOrPatch()}
        <h2>${this.props.payload.arch.title ?? "Map"}</h2>
        <p class="sum-map-hint">${count} located record(s).</p>
        <div
          class="sum-map-canvas"
          ref="map-canvas"
          style="height:480px;border-radius:8px;border:1px solid var(--sum-border,#ddd)"
        ></div>
      </div>
    `;
  }
}
