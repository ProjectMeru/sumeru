package server

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/module"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// Public URL paths for assets served outside the core /static/ file tree.
const (
	staticURLPrefix       = "/static/"
	moduleIconURLPrefix   = "/static/module-icon/"
	brandStylesheetURL    = "/static/brand.css"
	appLogoURL            = "/static/app-logo"
	addonThemeOverrideRel = "static/css/theme-overrides.css"
)

// layoutStylesheetCount is how many DefaultStylesheetURLs entries precede view CSS;
// sumeru_ai.css is inserted immediately after those layout slices when present.
const layoutStylesheetCount = 6

// registerBrandingAndStatic registers HTTP handlers and render hints for shell chrome:
//   - bundled core CSS (optionally with sumeru_ai)
//   - installed-addon theme override CSS
//   - optional brand CSS and app logo from config
//   - per-module icons from addon manifests
//   - the core asset directory at /static/ (registered last as a catch-all)
func registerBrandingAndStatic() {
	ctx := context.Background()

	stylesheetURLs := buildShellStylesheetURLs()
	scriptURLs := []string{}
	registerInstalledAddonThemeOverrides(ctx, &stylesheetURLs)
	registerInstalledAddonManifestAssets(ctx, &stylesheetURLs, &scriptURLs)
	registerInstalledAddonSwcEntries(ctx, &scriptURLs)
	registerBrandStylesheet(ctx, &stylesheetURLs)
	render.SetExtraStylesheetURLs(stylesheetURLs)
	render.SetExtraScriptURLs(scriptURLs)

	logoURL := registerAppLogo(ctx)
	render.SetShellBranding(render.ShellBranding{
		LogoURL: logoURL,
		Company: strings.TrimSpace(config.AppConfig.CompanyDisplayName),
		User:    strings.TrimSpace(config.AppConfig.UserDisplayName),
	})

	http.HandleFunc(moduleIconURLPrefix, serveModuleIcon)
	registerCoreAssetFileServer(ctx)
}

// buildShellStylesheetURLs returns the default shell CSS list, inserting the AI addon
// sheet after the layout slices when sumeru_ai is installed.
func buildShellStylesheetURLs() []string {
	urls := append([]string(nil), assets.DefaultStylesheetURLs()...)
	if !addonModuleInstalled("sumeru_ai") {
		return urls
	}
	withAI := make([]string, 0, len(urls)+1)
	withAI = append(withAI, urls[:layoutStylesheetCount]...)
	withAI = append(withAI, assets.AIStylesheetURL())
	withAI = append(withAI, urls[layoutStylesheetCount:]...)
	return withAI
}

// registerInstalledAddonThemeOverrides exposes optional static/css/theme-overrides.css
// from each installed addon at /static/addon-css/<module>.css.
func registerInstalledAddonThemeOverrides(ctx context.Context, stylesheetURLs *[]string) {
	for _, moduleName := range sortedLoadedAddonNames() {
		addon := module.LoadedAddons[moduleName]
		if addon == nil || !addonModuleInstalled(moduleName) {
			continue
		}
		overridePath := filepath.Join(addon.Path, addonThemeOverrideRel)
		if !isRegularFile(overridePath) {
			continue
		}
		publicURL := "/static/addon-css/" + moduleName + ".css"
		registerInstalledModuleFileHandler(publicURL, overridePath, moduleName, "text/css; charset=utf-8")
		*stylesheetURLs = append(*stylesheetURLs, publicURL)
		applog.InfoMsg(ctx, "web", "static", "Registered addon theme overrides",
			map[string]interface{}{"module": moduleName, "path": overridePath, "url": publicURL})
	}
}

// registerInstalledAddonManifestAssets exposes manifest.json "assets" entries at
// /static/addon-asset/<module>/<relative-path>. CSS files are appended to the shell stylesheet list;
// JS files are appended to ExtraScriptURLs.
func registerInstalledAddonManifestAssets(ctx context.Context, stylesheetURLs, scriptURLs *[]string) {
	for _, moduleName := range sortedLoadedAddonNames() {
		addon := module.LoadedAddons[moduleName]
		if addon == nil || !addonModuleInstalled(moduleName) {
			continue
		}
		for _, rel := range addon.Manifest.Assets {
			registerManifestAsset(ctx, addon, moduleName, rel, stylesheetURLs, scriptURLs)
		}
		for bundle, paths := range addon.Manifest.AssetBundles {
			for _, rel := range paths {
				registerManifestAsset(ctx, addon, moduleName, rel, stylesheetURLs, scriptURLs)
				applog.InfoMsg(ctx, "web", "static", "Registered manifest asset bundle entry",
					map[string]interface{}{"module": moduleName, "bundle": bundle, "path": rel})
			}
		}
	}
}

func registerInstalledAddonSwcEntries(ctx context.Context, scriptURLs *[]string) {
	for _, moduleName := range sortedLoadedAddonNames() {
		addon := module.LoadedAddons[moduleName]
		if addon == nil || !addonModuleInstalled(moduleName) {
			continue
		}
		entry := strings.TrimSpace(addon.Manifest.SwcEntry)
		if entry == "" {
			continue
		}
		clean, ok := normalizeManifestAssetRel(entry)
		if !ok {
			continue
		}
		absPath := filepath.Join(addon.Path, clean)
		if !isRegularFile(absPath) {
			applog.WarnMsg(ctx, "web", "static", "swc_entry not found", nil,
				map[string]interface{}{"module": moduleName, "path": absPath})
			continue
		}
		publicURL := manifestAssetPublicURL(moduleName, clean)
		registerInstalledModuleFileHandler(publicURL, absPath, moduleName, "text/javascript; charset=utf-8")
		*scriptURLs = append(*scriptURLs, publicURL)
		applog.InfoMsg(ctx, "web", "static", "Registered swc_entry",
			map[string]interface{}{"module": moduleName, "path": absPath, "url": publicURL})
	}
}

func registerManifestAsset(ctx context.Context, addon *module.Addon, moduleName, rel string, stylesheetURLs, scriptURLs *[]string) {
	clean, ok := normalizeManifestAssetRel(rel)
	if !ok {
		return
	}
	absPath := filepath.Join(addon.Path, clean)
	if !isRegularFile(absPath) {
		applog.WarnMsg(ctx, "web", "static", "manifest asset not found", nil,
			map[string]interface{}{"module": moduleName, "path": absPath})
		return
	}
	publicURL := manifestAssetPublicURL(moduleName, clean)
	contentType := contentTypeForAssetRel(clean)
	registerInstalledModuleFileHandler(publicURL, absPath, moduleName, contentType)
	if strings.HasSuffix(strings.ToLower(clean), ".css") {
		*stylesheetURLs = append(*stylesheetURLs, publicURL)
	}
	if strings.HasSuffix(strings.ToLower(clean), ".js") || strings.HasSuffix(strings.ToLower(clean), ".mjs") {
		*scriptURLs = append(*scriptURLs, publicURL)
	}
	applog.InfoMsg(ctx, "web", "static", "Registered manifest asset",
		map[string]interface{}{"module": moduleName, "path": absPath, "url": publicURL})
}

func normalizeManifestAssetRel(rel string) (string, bool) {
	rel = strings.TrimSpace(rel)
	if strings.Contains(rel, `\`) {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
		return "", false
	}
	if filepath.IsAbs(rel) {
		return "", false
	}
	if len(rel) >= 2 && rel[1] == ':' {
		return "", false
	}
	return rel, true
}

func manifestAssetPublicURL(moduleName, cleanRel string) string {
	return "/static/addon-asset/" + moduleName + "/" + cleanRel
}

func contentTypeForAssetRel(rel string) string {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func registerBrandStylesheet(ctx context.Context, stylesheetURLs *[]string) {
	brandPath := strings.TrimSpace(config.AppConfig.BrandCSS)
	if brandPath == "" {
		return
	}
	if !isRegularFile(brandPath) {
		applog.WarnMsg(ctx, "web", "static", "brand_css path not found or not a file", nil,
			map[string]interface{}{"path": brandPath})
		return
	}
	registerStaticFileHandler(brandStylesheetURL, brandPath, "text/css; charset=utf-8")
	*stylesheetURLs = append(*stylesheetURLs, brandStylesheetURL)
	applog.InfoMsg(ctx, "web", "static", "Registered brand stylesheet",
		map[string]interface{}{"path": brandPath})
}

// registerAppLogo serves config.AppConfig.LogoPath at /static/app-logo when the file exists.
func registerAppLogo(ctx context.Context) string {
	logoPath := strings.TrimSpace(config.AppConfig.LogoPath)
	if logoPath == "" {
		return ""
	}
	if !isRegularFile(logoPath) {
		applog.WarnMsg(ctx, "web", "static", "logo_path not found or not a file", nil,
			map[string]interface{}{"path": logoPath})
		return ""
	}
	http.HandleFunc(appLogoURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentTypeForImageExt(filepath.Ext(logoPath)))
		http.ServeFile(w, r, logoPath)
	})
	applog.InfoMsg(ctx, "web", "static", "Registered application logo",
		map[string]interface{}{"path": logoPath})
	return appLogoURL
}

// registerCoreAssetFileServer serves bundled JS/CSS/images from config.AppConfig.AssetsPath.
// Must run after specific /static/* handlers so those routes are not shadowed.
func registerCoreAssetFileServer(ctx context.Context) {
	assetsRoot := filepath.Clean(config.AppConfig.AssetsPath)
	fileServer := http.FileServer(http.Dir(assetsRoot))
	http.Handle(staticURLPrefix, http.StripPrefix(staticURLPrefix, fileServer))
	applog.InfoMsg(ctx, "web", "static", "Serving static files",
		map[string]interface{}{"path": assetsRoot})
}

// registerStaticFileHandler serves a fixed file at urlPath with a constant Content-Type.
func registerStaticFileHandler(urlPath, filePath, contentType string) {
	servePath := filePath
	http.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, servePath)
	})
}

// registerInstalledModuleFileHandler serves a file only while the owning module stays installed.
func registerInstalledModuleFileHandler(urlPath, filePath, moduleName, contentType string) {
	servePath := filePath
	moduleName = strings.TrimSpace(moduleName)
	http.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
		if !addonModuleInstalled(moduleName) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		http.ServeFile(w, r, servePath)
	})
}

func sortedLoadedAddonNames() []string {
	names := make([]string, 0, len(module.LoadedAddons))
	for name := range module.LoadedAddons {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func contentTypeForImageExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// addonModuleInstalled reports whether moduleName is active and in the installed state.
func addonModuleInstalled(moduleName string) bool {
	if orm.DB == nil {
		return false
	}
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return false
	}
	table := orm.MustQuotedTableName("sys.module")
	var state string
	err := orm.DB.QueryRow(
		`SELECT state FROM `+table+` WHERE name = $1 AND active = true`,
		moduleName,
	).Scan(&state)
	return err == nil && strings.TrimSpace(state) == "installed"
}

// serveModuleIcon serves GET /static/module-icon/<module> from the addon tree.
func serveModuleIcon(w http.ResponseWriter, r *http.Request) {
	moduleName := strings.Trim(strings.TrimPrefix(r.URL.Path, moduleIconURLPrefix), "/")
	if moduleName == "" || strings.Contains(moduleName, "/") {
		http.NotFound(w, r)
		return
	}
	if !addonModuleInstalled(moduleName) {
		http.NotFound(w, r)
		return
	}
	iconRelPath := moduleIconRelPathFromDB(r, moduleName)
	iconFile := render.ModuleIconServePath(moduleName, iconRelPath)
	if iconFile == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentTypeForImageExt(filepath.Ext(iconFile)))
	http.ServeFile(w, r, iconFile)
}

func moduleIconRelPathFromDB(r *http.Request, moduleName string) string {
	if orm.DB == nil {
		return ""
	}
	row, err := orm.SearchOne(r.Context(), "sys.module", map[string]interface{}{"name": moduleName})
	if err != nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(row["icon"]))
}
