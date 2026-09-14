package render

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/orm"
)

func ParsePinnedAppsJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

// SanitizePinnedModuleList drops unknown modules and duplicates; stable order.
func SanitizePinnedModuleList(raw []string, allowed map[string]struct{}) []string {
	if len(raw) == 0 || len(allowed) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, name := range raw {
		mod := strings.TrimSpace(name)
		if mod == "" || mod == "base" {
			continue
		}
		if _, ok := allowed[mod]; !ok {
			continue
		}
		if _, dup := seen[mod]; dup {
			continue
		}
		seen[mod] = struct{}{}
		out = append(out, mod)
	}
	return out
}

// AllowedPinModuleNames returns module technical names the user may pin (visible top-bar apps).
func AllowedPinModuleNames(ctx context.Context) map[string]struct{} {
	allowed := make(map[string]struct{})
	topMenus, _, _, _ := LoadShellMenus(ctx, "")
	for _, m := range topMenus {
		mod := strings.TrimSpace(m.Module)
		if mod == "" || mod == "base" {
			continue
		}
		allowed[mod] = struct{}{}
	}
	return allowed
}

// PinnedAppsForUser loads and sanitizes pinned modules for the signed-in user.
func PinnedAppsForUser(ctx context.Context) ([]string, error) {
	uid := orm.UIDFromContext(ctx)
	if uid <= 0 || orm.DB == nil {
		return nil, nil
	}
	row, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid})
	if err != nil {
		return nil, err
	}
	raw := ParsePinnedAppsJSON(orm.AsString(row["pinned_apps"]))
	allowed := AllowedPinModuleNames(ctx)
	return SanitizePinnedModuleList(raw, allowed), nil
}

// SavePinnedAppsForUser validates and persists pinned modules for the signed-in user.
func SavePinnedAppsForUser(ctx context.Context, modules []string) ([]string, error) {
	uid := orm.UIDFromContext(ctx)
	if uid <= 0 {
		return nil, fmt.Errorf("not authenticated")
	}
	allowed := AllowedPinModuleNames(ctx)
	clean := SanitizePinnedModuleList(modules, allowed)
	if clean == nil {
		clean = []string{}
	}
	b, err := json.Marshal(clean)
	if err != nil {
		return nil, err
	}
	if err := orm.UpdateRecordByID(ctx, "core.user", uid, map[string]interface{}{
		"pinned_apps": string(b),
	}); err != nil {
		return nil, err
	}
	return clean, nil
}

// BuildPinnedAppsJSON returns sanitized pinned modules for shell bootstrap.
func BuildPinnedAppsJSON(ctx context.Context) template.JS {
	mods, err := PinnedAppsForUser(ctx)
	if err != nil || len(mods) == 0 {
		return template.JS("[]")
	}
	b, err := json.Marshal(mods)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}
