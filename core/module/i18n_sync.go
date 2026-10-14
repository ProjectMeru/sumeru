package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/orm"
)

// SyncInstalledAddonPOFiles imports i18n/*.po for every installed addon on addons_path.
// Kernel, standard, and custom modules use the same folder. i18n must be installed first
// because it owns sys.translation.
func SyncInstalledAddonPOFiles(ctx context.Context) error {
	if orm.DB == nil {
		return nil
	}
	installed, err := orm.InstalledModuleNames(ctx)
	if err != nil {
		return err
	}
	if _, ok := installed["i18n"]; !ok {
		return nil
	}
	names := make([]string, 0, len(installed))
	for name := range installed {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		addon, ok := DiscoveredAddons[name]
		if !ok {
			continue
		}
		if err := SyncAddonTranslationPOFiles(ctx, addon.Path, name); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

// SyncAddonTranslationPOFiles imports gettext PO files from addonPath/i18n/*.po.
func SyncAddonTranslationPOFiles(ctx context.Context, addonPath, moduleName string) error {
	if orm.DB == nil {
		return nil
	}
	dir := filepath.Join(addonPath, "i18n")
	matches, err := filepath.Glob(filepath.Join(dir, "*.po"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return nil
	}
	// sys.translation is owned by i18n, which auto-installs after base. PO import
	// runs during base install, so the table and (lang, src, module) unique index
	// must exist before the upsert.
	if err := orm.SyncRegistrySchemaForNames([]string{"sys.translation"}); err != nil {
		return err
	}
	var firstErr error
	for _, poPath := range matches {
		f, err := os.Open(poPath)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		langHint, entries, err := orm.ParsePO(f)
		_ = f.Close()
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", filepath.Base(poPath), err)
			}
			continue
		}
		lang := strings.TrimSpace(langHint)
		if lang == "" {
			lang = orm.LangCodeFromPOFilename(poPath)
		}
		if _, err := orm.ImportTranslationsPO(ctx, orm.DB, moduleName, lang, entries); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", filepath.Base(poPath), err)
			}
		}
	}
	return firstErr
}
