package render

import (
	"context"
	"encoding/json"
	"html/template"
	"strings"

	"sumeru/addons/mail"
	"sumeru/core/orm"
)

// SWCBootstrapWorkspace is the initial workspace route embedded in bootstrap JSON.
type SWCBootstrapWorkspace struct {
	ActionID   int    `json:"actionId"`
	MenuID     string `json:"menuId"`
	ViewType   string `json:"viewType"`
	RecordID   int    `json:"recordId"`
	FormEdit   bool   `json:"formEdit"`
	ListSearch string `json:"listSearch"`
}

type swcBootstrap struct {
	CSRFToken           string                 `json:"csrfToken"`
	RPCURL              string                 `json:"rpcUrl"`
	SwcAPIBase          string                 `json:"swcApiBase"`
	User                swcBootstrapUser       `json:"user"`
	Company             swcBootstrapCompany    `json:"company"`
	Companies           []swcBootstrapCompany  `json:"companies"`
	ActiveCompanyID     int                    `json:"activeCompanyId"`
	ShowCompanySwitcher bool                   `json:"showCompanySwitcher"`
	TopMenus            []swcBootstrapMenu     `json:"topMenus"`
	SidebarMenus        []swcBootstrapSidebar  `json:"sidebarMenus"`
	ActiveModuleID      string                 `json:"activeModuleId"`
	ActiveMenuID        string                 `json:"activeMenuId"`
	Apps                []swcBootstrapApp      `json:"apps"`
	PinnedApps          []string               `json:"pinnedApps"`
	AppsNavAllowed      bool                   `json:"appsNavAllowed"`
	SettingsNavAllowed  bool                   `json:"settingsNavAllowed"`
	ActivityEnabled     bool                   `json:"activityEnabled"`
	BusEnabled          bool                   `json:"busEnabled"`
	DocsURL             string                 `json:"docsUrl"`
	ProfileURL          string                 `json:"profileUrl"`
	Features            map[string]bool        `json:"features,omitempty"`
	Workspace           *SWCBootstrapWorkspace `json:"workspace,omitempty"`
	Toasts              []swcBootstrapToast    `json:"toasts,omitempty"`
}

type swcBootstrapUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Login    string `json:"login"`
	Image    string `json:"image,omitempty"`
	Initials string `json:"initials"`
}

type swcBootstrapCompany struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type swcBootstrapMenu struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Action  string `json:"action"`
	Module  string `json:"module,omitempty"`
	WebIcon string `json:"webIcon,omitempty"`
}

type swcBootstrapSidebar struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Sequence int                `json:"sequence"`
	SubMenus []swcBootstrapMenu `json:"subMenus"`
}

type swcBootstrapApp struct {
	Kind        string `json:"kind,omitempty"`
	Module      string `json:"module"`
	Name        string `json:"name"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
}

type swcBootstrapToast struct {
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Details string `json:"details,omitempty"`
}

// BuildSWCBootstrapJSON builds window.__SWC_BOOTSTRAP__ for base.html.
func BuildSWCBootstrapJSON(ctx context.Context, page PageData, ws *SWCBootstrapWorkspace) template.JS {
	topMenus, sidebarMenus, activeModuleID, _ := LoadShellMenus(ctx, page.ActiveMenuID)
	b := swcBootstrap{
		CSRFToken:           page.CSRFToken,
		RPCURL:              "/api/rpc",
		SwcAPIBase:          "/web/swc",
		ActiveModuleID:      activeModuleID,
		ActiveMenuID:        page.ActiveMenuID,
		AppsNavAllowed:      page.AppsNavAllowed,
		SettingsNavAllowed:  page.SettingsNavAllowed,
		ActivityEnabled:     shellActivityEnabled(ctx, page),
		BusEnabled:          true,
		DocsURL:             page.UserDocsHref,
		ProfileURL:          page.UserProfileHref,
		Features:            buildSWCFeatureFlags(ctx),
		Workspace:           ws,
		ShowCompanySwitcher: page.ShowCompanySwitcher,
		ActiveCompanyID:     page.ShellActiveCompanyID,
	}
	for _, m := range topMenus {
		b.TopMenus = append(b.TopMenus, swcBootstrapMenu{
			ID: m.ID, Name: m.Name, Action: m.Action, Module: m.Module, WebIcon: m.WebIcon,
		})
	}
	for _, g := range sidebarMenus {
		sg := swcBootstrapSidebar{ID: g.ID, Name: g.Name, Sequence: g.Sequence}
		for _, sm := range g.SubMenus {
			sg.SubMenus = append(sg.SubMenus, swcBootstrapMenu{
				ID: sm.ID, Name: sm.Name, Action: sm.Action, Module: sm.Module, WebIcon: sm.WebIcon,
			})
		}
		b.SidebarMenus = append(b.SidebarMenus, sg)
	}
	for _, c := range page.ShellCompanyOptions {
		b.Companies = append(b.Companies, swcBootstrapCompany{ID: c.ID, Name: c.Name})
	}
	b.Company = swcBootstrapCompany{Name: page.ShellCompany}
	if page.ShellActiveCompanyID > 0 {
		for _, c := range page.ShellCompanyOptions {
			if c.ID == page.ShellActiveCompanyID {
				b.Company = swcBootstrapCompany{ID: c.ID, Name: c.Name}
				break
			}
		}
	}
	uid := orm.UIDFromContext(ctx)
	if uid > 0 {
		if u, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid}); err == nil {
			b.User = swcBootstrapUser{
				ID:       uid,
				Name:     strings.TrimSpace(orm.AsString(u["name"])),
				Login:    strings.TrimSpace(orm.AsString(u["login"])),
				Image:    strings.TrimSpace(orm.AsString(u["image"])),
				Initials: UserInitialsFromName(strings.TrimSpace(orm.AsString(u["name"]))),
			}
		}
	}
	b.Apps = parseSWCLauncherApps(page.AppLauncherJSON)
	pinnedJSON := page.PinnedAppsJSON
	if len(pinnedJSON) == 0 {
		pinnedJSON = BuildPinnedAppsJSON(ctx)
	}
	b.PinnedApps = parseSWCPinnedApps(pinnedJSON)
	for _, t := range page.ToastMessages {
		b.Toasts = append(b.Toasts, swcBootstrapToast{
			Kind: t.Kind, Title: t.Title, Body: t.Body, Details: t.Details,
		})
	}
	raw, err := json.Marshal(b)
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(raw)
}

func parseSWCLauncherApps(raw template.JS) []swcBootstrapApp {
	var items []struct {
		Kind        string `json:"kind"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
		Href        string `json:"href"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	out := make([]swcBootstrapApp, 0, len(items))
	for _, it := range items {
		kind := strings.TrimSpace(it.Kind)
		if kind == "" {
			kind = "app"
		}
		out = append(out, swcBootstrapApp{
			Kind:        kind,
			Module:      it.Name,
			Name:        it.DisplayName,
			Action:      it.Href,
			Description: strings.TrimSpace(it.Description),
		})
	}
	return out
}

func shellActivityEnabled(ctx context.Context, page PageData) bool {
	if page.SuppressActivityDock {
		return false
	}
	if page.ActivityEnabled {
		return true
	}
	return mail.CompanyChatterEnabled(ctx) && mail.CompanyActivityPanelEnabled(ctx)
}

func parseSWCPinnedApps(raw template.JS) []string {
	var out []string
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func buildSWCFeatureFlags(ctx context.Context) map[string]bool {
	uid := orm.UIDFromContext(ctx)
	if uid <= 0 {
		return nil
	}
	flags := map[string]bool{}
	if orm.UserHasAnyAccessGroup(ctx, uid, "studio.group_studio_user") {
		flags["studio"] = true
	}
	if len(flags) == 0 {
		return nil
	}
	return flags
}
