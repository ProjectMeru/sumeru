package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func processXMLRecords(ctx context.Context, moduleName string, records []parser.Record, inheritQueue *[]parser.Record, opts dataFileOpts) {
	for _, xmlRecord := range records {
		if opts.skipExistingOnUpdate(ctx, moduleName, xmlRecord.ID) {
			continue
		}
		if xmlRecord.Model == "sys.action.window" {
			upsertSysActionWindowFromRecord(ctx, moduleName, xmlRecord)
		}
		if xmlRecord.Model == "sys.action.url" {
			upsertSysActionURLFromRecord(ctx, moduleName, xmlRecord)
		}
		if xmlRecord.Model == "sys.view" {
			if strings.TrimSpace(parser.RecordFieldMap(xmlRecord)["inherit_id"]) != "" {
				*inheritQueue = append(*inheritQueue, xmlRecord)
			} else {
				upsertSysViewFromRecord(ctx, moduleName, xmlRecord)
			}
		}
		syncGenericRegistryRecord(ctx, moduleName, xmlRecord)
	}
}

func RecordsFromActions(actions []parser.Action) []parser.Record {
	if len(actions) == 0 {
		return nil
	}
	out := make([]parser.Record, 0, len(actions))
	for _, a := range actions {
		out = append(out, a.ToRecord())
	}
	return out
}

type parsedManifestFile struct {
	noUpdate  bool
	records   []parser.Record
	views     []parser.View
	menuItems []parser.MenuItem
}

func (p parsedManifestFile) hasContent() bool {
	return len(p.records) > 0 || len(p.views) > 0 || len(p.menuItems) > 0
}

func manifestFromViewList(xmlPath string) (parsedManifestFile, error) {
	parsedViewData, err := parser.ParseViewList(xmlPath)
	if err != nil {
		return parsedManifestFile{}, err
	}
	out := parsedManifestFile{
		noUpdate:  parsedViewData.NoUpdate,
		records:   append([]parser.Record(nil), parsedViewData.Records...),
		views:     parsedViewData.Views,
		menuItems: parsedViewData.MenuItems,
	}
	out.records = append(out.records, RecordsFromActions(parsedViewData.Actions)...)
	return out, nil
}

func manifestFromMenuList(xmlPath string) (parsedManifestFile, error) {
	menuList, err := parser.ParseMenuList(xmlPath)
	if err != nil {
		return parsedManifestFile{}, err
	}
	out := parsedManifestFile{
		noUpdate:  menuList.NoUpdate,
		records:   append([]parser.Record(nil), menuList.Records...),
		menuItems: menuList.MenuItems,
	}
	out.records = append(out.records, RecordsFromActions(menuList.Actions)...)
	return out, nil
}

func resolveManifestFile(xmlPath string, acceptView func(parsedManifestFile) bool) (parsedManifestFile, error) {
	viewParsed, viewErr := manifestFromViewList(xmlPath)
	if viewErr == nil && acceptView(viewParsed) {
		return viewParsed, nil
	}
	if viewErr != nil && !AllowMenuParseFallback(viewErr) {
		return parsedManifestFile{}, viewErr
	}

	menuParsed, menuErr := manifestFromMenuList(xmlPath)
	if menuErr == nil {
		if menuParsed.hasContent() {
			return menuParsed, nil
		}
		return parsedManifestFile{}, nil
	}
	if viewErr != nil {
		return parsedManifestFile{}, fmt.Errorf("ParseViewList: %v; ParseMenuList: %v", viewErr, menuErr)
	}
	return parsedManifestFile{}, menuErr
}

func syncParsedManifestFile(ctx context.Context, moduleName string, parsed parsedManifestFile, inheritQueue *[]parser.Record, menuCollector *[]parser.MenuItem) {
	if len(parsed.records) > 0 {
		processXMLRecords(ctx, moduleName, parsed.records, inheritQueue, dataFileOpts{noUpdate: parsed.noUpdate})
	}
	for i := range parsed.views {
		upsertInlineViewDef(ctx, moduleName, &parsed.views[i])
	}
	if len(parsed.menuItems) > 0 {
		*menuCollector = append(*menuCollector, parsed.menuItems...)
	}
}

// loadManifestDataFile parses one manifest XML path (view or menu layout) and syncs its content.
// Menu items are appended to menuCollector for a deferred sync pass after all manifest files load.
func loadManifestDataFile(ctx context.Context, moduleName, xmlPath, relFile string, inheritQueue *[]parser.Record, menuCollector *[]parser.MenuItem) []error {
	parsed, err := resolveManifestFile(xmlPath, func(p parsedManifestFile) bool { return p.hasContent() })
	if err != nil {
		return []error{RecoverableSync(moduleName, "parse "+relFile, err)}
	}
	if !parsed.hasContent() {
		return nil
	}
	syncParsedManifestFile(ctx, moduleName, parsed, inheritQueue, menuCollector)
	return nil
}

// CollectMenuItemsFromManifestFile parses menu items from one manifest XML path (for tests and tooling).
func CollectMenuItemsFromManifestFile(xmlPath string) ([]parser.MenuItem, error) {
	parsed, err := resolveManifestFile(xmlPath, func(p parsedManifestFile) bool { return len(p.menuItems) > 0 })
	if err != nil {
		return nil, err
	}
	return append([]parser.MenuItem(nil), parsed.menuItems...), nil
}

func AllowMenuParseFallback(viewErr error) bool {
	if viewErr == nil {
		return true
	}
	msg := viewErr.Error()
	// Invalid module root will fail menu parse the same way — do not double-warn.
	if strings.Contains(msg, "module XML root must be") {
		return false
	}
	return true
}

func (addon *Addon) SyncToDB(ctx context.Context) error {
	moduleName := addon.Manifest.Name
	var errs []error

	for _, registeredModel := range orm.Registry {
		if strings.TrimSpace(orm.DeclaringModule(registeredModel.ModelName())) != moduleName {
			continue
		}
		_, err := orm.Upsert(ctx, orm.RegistryModel("sys.model"), map[string]interface{}{
			"name":   registeredModel.ModelName(),
			"model":  registeredModel.ModelName(),
			"module": moduleName,
		}, "name")
		if err != nil {
			errs = append(errs, FatalSync(moduleName, "sys.model upsert "+registeredModel.ModelName(), err))
		}
	}

	if err := addon.syncCSVModelAccess(ctx); err != nil {
		errs = append(errs, FatalSync(moduleName, "CSV ACL load", err))
	} else {
		orm.InvalidateRuleCache()
	}
	var inheritQueue []parser.Record
	var deferredMenus []parser.MenuItem

	for _, xmlFile := range addon.Manifest.Data {
		if strings.HasSuffix(strings.ToLower(strings.TrimSpace(xmlFile)), ".csv") {
			continue // ACL CSV is loaded by syncCSVModelAccess above
		}
		xmlPath := filepath.Join(addon.Path, xmlFile)
		if _, err := os.Stat(xmlPath); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "data file missing "+xmlFile, err))
			continue
		}

		if fileErrs := loadManifestDataFile(ctx, moduleName, xmlPath, xmlFile, &inheritQueue, &deferredMenus); len(fileErrs) > 0 {
			errs = append(errs, fileErrs...)
		}
	}

	if len(deferredMenus) > 0 {
		syncMenusFromItems(ctx, moduleName, deferredMenus)
	}

	for _, xmlRecord := range inheritQueue {
		if err := applySysUIViewInherit(ctx, moduleName, xmlRecord); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "view inherit "+xmlRecord.ID, err))
		}
	}

	return aggregateErrors(moduleName, errs)
}
