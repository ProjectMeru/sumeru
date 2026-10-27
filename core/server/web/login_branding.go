package web

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const (
	defaultSumeruLogoURL      = "/static/img/sumeru-logo.svg"
	loginBrandLogoRoute       = "/web/login/logo"
	loginBrandPlatformName    = "Sumeru"
	defaultLoginHeadline      = "Enterprise workspace"
	defaultLoginTagline       = "Role-based access, audit trails, and optional single sign-on for your organization."
	defaultLoginTenantTagline = "Sign in to continue to your workspace."
)

// loginBrand is the left-panel hero on auth pages (login, TOTP, setup).
type loginBrand struct {
	LogoURL        string
	LogoAlt        string
	Kicker         string
	Headline       string
	Tagline        string
	ShowTrustChips bool
	ShowPoweredBy  bool
	AccentClass    string
}

type loginCompanyRow struct {
	Name         string
	LoginLogo    string
	LoginTagline string
	Color        int
}

var kanbanAccentHex = []string{
	"#875a7b", "#9c9c9c", "#67917a", "#a06666", "#8899aa", "#d4a017",
	"#6a5acd", "#4682b4", "#c0392b", "#e67e22", "#27ae60", "#3498db",
}

func resolveLoginBrand(ctx context.Context) loginBrand {
	row := loadLoginCompanyRow(ctx)
	return buildLoginBrand(row)
}

func defaultSumeruLoginBrand() loginBrand {
	return loginBrand{
		LogoURL:        defaultSumeruLogoURL,
		LogoAlt:        loginBrandPlatformName,
		Kicker:         loginBrandPlatformName,
		Headline:       defaultLoginHeadline,
		Tagline:        defaultLoginTagline,
		ShowTrustChips: true,
	}
}

func buildLoginBrand(row loginCompanyRow) loginBrand {
	out := defaultSumeruLoginBrand()

	cfgName := strings.TrimSpace(config.AppConfig.CompanyDisplayName)
	companyName := strings.TrimSpace(row.Name)
	if companyName == "" {
		companyName = cfgName
	}

	tagline := strings.TrimSpace(row.LoginTagline)
	hasCompanyLogo := strings.TrimSpace(row.LoginLogo) != ""
	shellLogo := strings.TrimSpace(render.ShellLogoURL())

	switch {
	case hasCompanyLogo:
		out.LogoURL = loginBrandLogoRoute
		out.LogoAlt = companyName
		if companyName == "" {
			out.LogoAlt = loginBrandPlatformName
		}
	case shellLogo != "":
		out.LogoURL = shellLogo
		out.LogoAlt = loginBrandPlatformName
	}

	tenantBranded := hasCompanyLogo || companyName != "" || tagline != ""
	if tenantBranded && companyName != "" {
		out.Kicker = "Welcome"
		out.Headline = companyName
		if tagline != "" {
			out.Tagline = tagline
		} else {
			out.Tagline = defaultLoginTenantTagline
		}
		out.ShowPoweredBy = true
	} else if cfgName != "" {
		out.Kicker = "Welcome"
		out.Headline = cfgName
		out.Tagline = defaultLoginTenantTagline
		out.ShowPoweredBy = true
	}

	if class := kanbanAccentClass(row.Color); class != "" {
		out.AccentClass = class
	}
	return out
}

func loadLoginCompanyRow(ctx context.Context) loginCompanyRow {
	ctx = orm.AuditedBypass(ctx, "login.brand")
	rows, err := orm.SearchLimit(ctx, "core.company", nil, 1)
	if err != nil {
		return loginCompanyRow{}
	}
	if 0 == len(rows) {
		return loginCompanyRow{}
	}

	rec := rows[0]
	return loginCompanyRow{
		Name:         oauthRowString(rec, "name"),
		LoginLogo:    oauthRowString(rec, "login_logo"),
		LoginTagline: oauthRowString(rec, "login_tagline"),
		Color:        int(oauthRowInt(rec, "color")),
	}
}

func loadLoginCompanyLogoPayload(ctx context.Context) string {
	return strings.TrimSpace(loadLoginCompanyRow(ctx).LoginLogo)
}

func kanbanAccentClass(colorIndex int) string {
	if colorIndex < 0 || colorIndex >= len(kanbanAccentHex) {
		return ""
	}
	return "sum-login-brand--accent-" + strconv.Itoa(colorIndex)
}

func totpLoginBrand(base loginBrand) loginBrand {
	base.Headline = "Verify your identity"
	base.Tagline = "Enter the code from your authenticator app to finish signing in."
	base.ShowTrustChips = false
	return base
}

func setupLoginBrand() loginBrand {
	b := defaultSumeruLoginBrand()
	b.Headline = "Database setup"
	b.Tagline = "Connect your PostgreSQL database, create your company, and define the first administrator account."
	b.ShowTrustChips = false
	return b
}
