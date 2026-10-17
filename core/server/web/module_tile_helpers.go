package web

import (
	"context"
	"sort"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/module"
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
		openHref := render.ModuleWorkspaceEntryURL(ctx, parsed.Name)
		_ = menuID
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

