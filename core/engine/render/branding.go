package render

import (
	"context"
	"html/template"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// AppDisplayName is the product name shown in the shell, document title, and manifests.
const AppDisplayName = "Sumeru"
const AppDocumentationURL = "https://projectmeru.github.io/sumeru/docs/"

// ShellBranding is header chrome (logo + labels) set once at process start from config.
type ShellBranding struct {
	LogoURL string // e.g. "/static/app-logo" or empty
	Company string
	User    string
}

var shell ShellBranding

// SetShellBranding configures header logo URL and optional company / user labels.
func SetShellBranding(b ShellBranding) {
	shell = b
}

// ShellLogoURL returns the configured shell logo URL (e.g. "/static/app-logo"), or empty.
func ShellLogoURL() string {
	return strings.TrimSpace(shell.LogoURL)
}

// EnrichShellPageData merges global shell branding and optional DB labels into page data.
func EnrichShellPageData(ctx context.Context, d *PageData) {
	d.AppName = AppDisplayName
	d.LogoURL = shell.LogoURL
	uidShell := orm.UIDFromContext(ctx)
	d.AppsNavAllowed = orm.UserHasGroupXML(ctx, uidShell, "base.group_system")
	d.SettingsNavAllowed = orm.UserHasGroupXML(ctx, uidShell, "base.group_user")
	if strings.TrimSpace(d.BrandLockupHref) == "" {
		d.BrandLockupHref = HomeWebURL()
	}
	if strings.TrimSpace(d.HomeNavHref) == "" {
		d.HomeNavHref = d.BrandLockupHref
	}
	if 0 == len(d.AppLauncherJSON) {
		d.AppLauncherJSON = BuildAppLauncherJSON(ctx)
	}
	if 0 == len(d.PinnedAppsJSON) {
		d.PinnedAppsJSON = BuildPinnedAppsJSON(ctx)
	}
	if strings.TrimSpace(d.UserProfileHref) == "" {
		d.UserProfileHref = "/web/settings"
	}
	if strings.TrimSpace(d.UserDocsHref) == "" {
		d.UserDocsHref = AppDocumentationURL + "using/index.html"
	}
	d.ShellCompany = strings.TrimSpace(shell.Company)
	d.ShellUser = strings.TrimSpace(shell.User)
	d.ShellUserImage = ""
	if orm.DB == nil {
		d.ShellUserInitials = UserInitialsFromName(d.ShellUser)
		return
	}

	d.ShellCompanyOptions = loadShellCompanies(ctx)
	d.ShowCompanySwitcher = len(d.ShellCompanyOptions) > 1

	uid := orm.UIDFromContext(ctx)
	activeCID := 0
	if uid > 0 {
		if u, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid}); err == nil {
			d.ShellUserLogin = strings.TrimSpace(orm.AsString(u["login"]))
			name := strings.TrimSpace(orm.AsString(u["name"]))
			if name == "" {
				name = d.ShellUserLogin
			}
			if name != "" {
				d.ShellUser = name
			}
			if img := strings.TrimSpace(orm.AsString(u["image"])); SafeImageSrc(img) {
				d.ShellUserImage = template.URL(img)
				cropRaw := strings.TrimSpace(orm.AsString(u["image_crop"]))
				if crop, ok := ParseImageCrop(cropRaw); ok {
					d.ShellUserImageCrop = AvatarCropStyle(crop, true)
				}
			}
			if id, ok := orm.CoerceInt64(u["company_id"]); ok && id > 0 {
				activeCID = int(id)
			}
		}
	}
	d.ShellActiveCompanyID = activeCID

	if d.ShellCompany == "" {
		d.ShellCompany = shellCompanyNameFromOptions(d.ShellCompanyOptions, activeCID)
	}
	if d.ShellUser == "" && config.AppConfig.DevMode {
		if _, ok := orm.Registry["core.user"]; ok {
			tn := orm.MustQuotedTableName("core.user")
			var nm string
			q := `SELECT COALESCE(NULLIF(TRIM(name), ''), TRIM(login), '') FROM ` + tn + ` ORDER BY id ASC LIMIT 1`
			if err := orm.DB.QueryRowContext(ctx, q).Scan(&nm); err == nil && strings.TrimSpace(nm) != "" {
				d.ShellUser = strings.TrimSpace(nm)
			}
		}
	}
	d.ShellUserInitials = UserInitialsFromName(d.ShellUser)

	// Activity dock (Messages + Log) stays visible for logged-in users; IM APIs gate chat separately.
	d.ActivityEnabled = orm.SecurityUID(ctx) > 0
	if 0 == len(d.ExtraScriptURLs) {
		d.ExtraScriptURLs = ExtraScriptURLs
	}
	if d.SWCEnabled && 0 == len(d.SwcAddonEntriesJSON) {
		d.SwcAddonEntriesJSON = BuildSwcAddonEntriesJSON()
	}
	if d.SuppressActivityDock {
		d.ActivityEnabled = false
	}
	d.appendShellHooks(ctx)
}

func (d *PageData) appendShellHooks(ctx context.Context) {
	var shellExtra strings.Builder
	for _, hook := range ShellHooks {
		shellExtra.WriteString(string(hook(ctx, nil, false)))
	}
	d.ShellExtraHTML = template.HTML(shellExtra.String())
}

func loadShellCompanies(ctx context.Context) []ShellCompanyOption {
	uid := orm.UIDFromContext(ctx)
	if 0 >= uid || orm.DB == nil {
		return nil
	}
	if _, ok := orm.Registry["core.company"]; !ok {
		return nil
	}
	allowed, err := orm.UserCompanyIDs(ctx, uid)
	if err != nil {
		return nil
	}
	if 0 == len(allowed) {
		return nil
	}

	allowedSet := make(map[int64]struct{}, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = struct{}{}
	}
	tn := orm.MustQuotedTableName("core.company")
	rows, err := orm.DB.QueryContext(ctx, `SELECT id, name FROM `+tn+` ORDER BY id ASC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ShellCompanyOption
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		name = strings.TrimSpace(name)
		if _, ok := allowedSet[int64(id)]; !ok || 0 >= id || name == "" {
			continue
		}
		out = append(out, ShellCompanyOption{ID: id, Name: name})
	}
	return out
}

func shellCompanyNameFromOptions(opts []ShellCompanyOption, preferID int) string {
	if preferID > 0 {
		for _, o := range opts {
			if o.ID == preferID {
				return o.Name
			}
		}
	}
	if len(opts) > 0 {
		return opts[0].Name
	}
	return ""
}
