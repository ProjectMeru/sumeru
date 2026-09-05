package module

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// ponytail: 16 rounds cover any sane menu depth; raise if addons nest deeper.
const maxMenuParentRounds = 16

func syncMenusFromItems(ctx context.Context, moduleName string, menus []parser.MenuItem) {
	buildValues := func(menu parser.MenuItem) map[string]interface{} {
		menuValues := map[string]interface{}{
			"name":          menu.Name,
			"sequence":      menu.Sequence,
			"module":        moduleName,
			"access_groups": strings.TrimSpace(menu.AccessGroups),
		}
		if menu.Action != "" {
			menuValues["action"] = menu.Action
			actionID, err := resolveXMLIDInModule(ctx, moduleName, menu.Action)
			if err == nil && actionID != 0 {
				menuValues["action_id"] = actionID
			} else {
				syncWarn(ctx, "Warning: sys.menu %s.%s action %q unresolved: %v", moduleName, menu.ID, menu.Action, err)
			}
		}
		if sanitizedIcon := sanitizeWebIcon(menu.WebIcon); sanitizedIcon != "" {
			menuValues["web_icon"] = sanitizedIcon
		}
		return menuValues
	}

	// Pass 1: application / section roots only (no parent in XML) → top bar candidates.
	for _, menu := range menus {
		if strings.TrimSpace(menu.ID) == "" || strings.TrimSpace(menu.ParentID) != "" {
			continue
		}
		menuValues := buildValues(menu)
		menuValues["parent_id"] = nil
		if _, err := upsertMenuRow(ctx, moduleName, menu.ID, menuValues); err != nil {
			syncWarn(ctx, "Warning: sys.menu root %s.%s: %v", moduleName, menu.ID, err)
		}
	}

	// Pass 2+: children. Fail closed — never persist a child with NULL parent_id (fake top-bar root).
	// Repeat until no progress so grandparents → parents → leaves work in one sync.
	pending := make([]parser.MenuItem, 0, len(menus))
	for _, menu := range menus {
		if strings.TrimSpace(menu.ID) == "" || strings.TrimSpace(menu.ParentID) == "" {
			continue
		}
		pending = append(pending, menu)
	}
	for round := 0; round < maxMenuParentRounds && len(pending) > 0; round++ {
		var next []parser.MenuItem
		progress := 0
		for _, menu := range pending {
			pid := strings.TrimSpace(menu.ParentID)
			parentID, err := resolveXMLIDInModule(ctx, moduleName, pid)
			if err != nil || parentID == 0 {
				next = append(next, menu)
				continue
			}
			menuValues := buildValues(menu)
			menuValues["parent_id"] = parentID
			if _, err := upsertMenuRow(ctx, moduleName, menu.ID, menuValues); err != nil {
				syncWarn(ctx, "Warning: sys.menu child %s.%s: %v", moduleName, menu.ID, err)
				next = append(next, menu)
				continue
			}
			progress++
		}
		pending = next
		if progress == 0 {
			break
		}
	}
	for _, menu := range pending {
		syncWarn(ctx, "Warning: sys.menu %s.%s parent %q unresolved — skipped (not creating top-bar orphan)",
			moduleName, menu.ID, strings.TrimSpace(menu.ParentID))
	}
}

func upsertMenuRow(ctx context.Context, moduleName, xmlID string, menuValues map[string]interface{}) (int, error) {
	rowID, _ := menuRowID(ctx, moduleName, xmlID)
	if rowID > 0 {
		if err := orm.UpdateRecordByID(ctx, "sys.menu", rowID, menuValues); err != nil {
			return 0, err
		}
	} else {
		id, err := orm.Create(ctx, orm.RegistryModel("sys.menu"), menuValues)
		if err != nil {
			return 0, err
		}
		rowID = id
	}
	if err := linkXMLRecord(ctx, moduleName, xmlID, "sys.menu", rowID); err != nil {
		return 0, err
	}
	return rowID, nil
}

func menuRowID(ctx context.Context, moduleName, xmlID string) (int, error) {
	md, err := orm.SearchOne(ctx, "sys.model.data", map[string]interface{}{
		"module": moduleName,
		"model":  "sys.menu",
		"name":   xmlID,
	})
	if err != nil {
		return 0, err
	}
	cid, ok := orm.CoerceInt64(md["core_id"])
	if !ok || cid <= 0 {
		return 0, fmt.Errorf("no core_id")
	}
	return int(cid), nil
}

func sanitizeWebIcon(iconString string) string {
	iconString = strings.TrimSpace(iconString)
	if iconString == "" {
		return ""
	}
	for _, char := range iconString {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return ""
		}
	}
	return iconString
}
