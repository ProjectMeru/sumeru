package web

import (
	"net/http"

	"sumeru/core/engine/render"
)

type homeDashData struct {
	Layout   string
	Tiles    []render.AppTile
	EmptyMsg string
}

// HomeDashboardHandler shows installed application modules as the signed-in user's app hub.
func HomeDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	layout := layoutFromQuery(r)

	tiles, err := loadInstalledAppTiles(ctx, true)
	if err != nil {
		WebLogEvent(ctx, WebLogInput{
			Route: r.URL.Path, Message: "Failed to list modules for home dashboard",
			Operation: "load", Status: "failure", Err: err,
		})
		http.Error(w, "Failed to list modules", http.StatusInternalServerError)
		return
	}

	tiles = filterHomeTiles(tiles)

	emptyMessage := ""
	if len(tiles) == 0 {
		emptyMessage = homeEmptyMessage
	}

	renderShellPage(w, r, shellPageOpts{
		Route:         r.URL.Path,
		InnerTemplate: homeInnerTemplate,
		InnerData: homeDashData{
			Layout:   layout,
			Tiles:    tiles,
			EmptyMsg: emptyMessage,
		},
		MenuIDStr: "",
		Page: render.PageData{
			Title:                homePageTitle,
			ViewBreadcrumb:       "Dashboard",
			ViewStylesheetURLs:   []string{homeStylesheetURL},
			SuppressActivityDock: true,
			SuppressSidebar:      true,
			ViewTabs:             render.HomeViewTabs(layout),
			BreadcrumbItems:      render.BuildHomeDashboardBreadcrumbs(ctx),
		},
	})
	WebLogNavigation(ctx, r.URL.Path, "module_hub", "Home dashboard opened", map[string]interface{}{
		"layout":     layout,
		"tile_count": len(tiles),
	})
}

func filterHomeTiles(tiles []render.AppTile) []render.AppTile {
	filtered := make([]render.AppTile, 0, len(tiles))
	for _, tile := range tiles {
		if tile.Name == baseModuleName {
			continue
		}
		if tile.IconURL == "" {
			tile.IconURL = render.DefaultAppIconURL
		}
		filtered = append(filtered, tile)
	}
	return filtered
}
