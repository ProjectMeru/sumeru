package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"sumeru/addons/mail"
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
		d.BrandLockupHref = HomeWebURL(ctx)
	}
	if strings.TrimSpace(d.BrandLockupHref) == "" {
		d.BrandLockupHref = "/web/home"
	}
	if strings.TrimSpace(d.HomeNavHref) == "" {
		d.HomeNavHref = d.BrandLockupHref
	}
	if len(d.AppLauncherJSON) == 0 {
		d.AppLauncherJSON = BuildAppLauncherJSON(ctx)
	}
	if len(d.PinnedAppsJSON) == 0 {
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
		d.applyShellUserInitials()
		return
	}

	d.ShellCompanyOptions = loadShellCompanies(ctx)
	d.ShowCompanySwitcher = len(d.ShellCompanyOptions) > 1

	uid := orm.UIDFromContext(ctx)
	activeCID := 0
	if uid > 0 {
		if u, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid}); err == nil {
			// Prefer the logged-in user's name and photo over static config labels.
			name := strings.TrimSpace(orm.AsString(u["name"]))
			if name == "" {
				name = strings.TrimSpace(orm.AsString(u["login"]))
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
	d.applyShellUserInitials()

	d.ActivityEnabled = mail.CompanyChatterEnabled(ctx) && mail.CompanyActivityPanelEnabled(ctx)
	if len(d.ExtraScriptURLs) == 0 {
		d.ExtraScriptURLs = ExtraScriptURLs
	}
	if d.SuppressActivityDock {
		d.ActivityEnabled = false
	}
	if !d.ActivityEnabled {
		d.ActivityLogItems = nil
		d.appendShellHooks(ctx)
		return
	}
	rows, err := mail.QueryActivityLog(ctx, 40, d.ActivityContextModel, d.ActivityContextRecordID)
	if err != nil {
		d.ActivityLogItems = nil
		d.appendShellHooks(ctx)
		return
	}
	for _, r := range rows {
		author := strings.TrimSpace(r.Author)
		if author == "" {
			author = "System"
		}
		meta := author
		if !r.CreateDate.IsZero() {
			meta = fmt.Sprintf("%s · %s", author, shortRelTime(r.CreateDate))
		}
		d.ActivityLogItems = append(d.ActivityLogItems, ActivityItem{Meta: meta, Body: strings.TrimSpace(r.Body)})
	}

	d.appendShellHooks(ctx)
}

func (d *PageData) applyShellUserInitials() {
	d.ShellUserInitials = UserInitialsFromName(d.ShellUser)
	if d.UserInitial == "" && d.ShellUser != "" {
		r := []rune(d.ShellUser)
		if len(r) > 0 {
			d.UserInitial = strings.ToUpper(string(r[0]))
		}
	}
}

func (d *PageData) appendShellHooks(ctx context.Context) {
	var shellExtra strings.Builder
	for _, hook := range ShellHooks {
		shellExtra.WriteString(string(hook(ctx, nil, false)))
	}
	d.ShellExtraHTML = template.HTML(shellExtra.String())
}

func loadShellCompanies(ctx context.Context) []ShellCompanyOption {
	if orm.DB == nil {
		return nil
	}
	if _, ok := orm.Registry["core.company"]; !ok {
		return nil
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
		if id > 0 && name != "" {
			out = append(out, ShellCompanyOption{ID: id, Name: name})
		}
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

func shortRelTime(t time.Time) string {
	t = t.UTC()
	now := time.Now().UTC()
	if t.After(now) {
		t = now
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	default:
		return t.Local().Format("Jan 02")
	}
}
