package render

import "strings"

const DefaultAppIconURL = "/static/img/projectmeru_icon.png"

type AppTile struct {
	Name         string
	DisplayName  string
	Version      string
	Description  string
	Author       string
	IconLetter   string
	IconHue      int
	IconURL      string
	WebIcon      string
	OpenMenuHref string
}

// IconHueFromString derives a stable hue from a module technical name.
func IconHueFromString(name string) int {
	hue := 265
	for _, c := range strings.TrimSpace(name) {
		hue = (hue*31 + int(c)) % 360
		if hue < 0 {
			hue += 360
		}
	}
	return hue
}
