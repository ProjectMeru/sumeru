package web

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/module"
	"sumeru/core/orm"
)

// loadInstalledAppTiles returns installed, active application modules sorted by display name.
// When forHome is true, sprite WebIcon is omitted (home uses static PNG icons only).
func loadInstalledAppTiles(ctx context.Context, forHome bool) ([]render.AppTile, error) {
	moduleRows, err := module.ListModules(ctx)
	if err != nil {
		return nil, err
	}
	var tiles []render.AppTile
	for _, row := range moduleRows {
		parsed, ok := parseModuleRow(row)
		if !ok || !parsed.Application || parsed.State != "installed" || !parsed.Active {
			continue
		}
		menuID, rootWebIcon := render.ModuleRootMenu(ctx, parsed.Name)
		openHref := appsRoute
		if menuID > 0 {
			openHref = menuHrefFromMenuID(menuID)
		}
		iconURL := render.ModuleIconURL(ctx, parsed.Name)
		webIcon := ""
		if !forHome && iconURL == "" {
			webIcon = rootWebIcon
		}
		tiles = append(tiles, render.AppTile{
			Name:         parsed.Name,
			DisplayName:  parsed.DisplayName,
			Version:      parsed.Version,
			Description:  parsed.Description,
			Author:       parsed.Author,
			IconLetter:   render.IconLetterFromName(parsed.DisplayName),
			IconHue:      render.IconHueFromString(parsed.Name),
			IconURL:      iconURL,
			WebIcon:      webIcon,
			OpenMenuHref: openHref,
		})
	}
	sort.Slice(tiles, func(i, j int) bool {
		displayI := strings.ToLower(strings.TrimSpace(tiles[i].DisplayName))
		displayJ := strings.ToLower(strings.TrimSpace(tiles[j].DisplayName))
		if displayI != displayJ {
			return displayI < displayJ
		}
		return tiles[i].Name < tiles[j].Name
	})
	return tiles, nil
}

// menuHrefFromMenuID builds a workspace URL for a numeric menu id.
func menuHrefFromMenuID(menuID int) string {
	return fmt.Sprintf("/web?menu_id=%d", menuID)
}

// menuHrefFromXMLID resolves a menu XML id to a workspace URL, or "" when not found.
func menuHrefFromXMLID(ctx context.Context, menuXMLID string) string {
	menuID, _, err := orm.ResolveXmlId(ctx, menuXMLID)
	if err != nil || menuID <= 0 {
		return ""
	}
	return menuHrefFromMenuID(menuID)
}
