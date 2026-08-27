package render

import (
	"context"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/swcmeta"
)

// UIHook allows addons to inject custom HTML into specific parts of the UI.
type UIHook func(ctx context.Context, vr *ViewRecordData, ro bool) template.HTML

var (
	// NotebookHooks: model -> page_title -> hook
	NotebookHooks = map[string]map[string]UIHook{}

	// ShellHooks are called for every page render to inject shell-level HTML.
	ShellHooks []UIHook
)

// RegisterShellHook adds a global hook to the shell rendering pipeline.
func RegisterShellHook(hook UIHook) {
	ShellHooks = append(ShellHooks, hook)
}

// RegisterNotebookHook registers a hook for a notebook page title on a model.
func RegisterNotebookHook(model, pageTitle string, hook UIHook) {
	if NotebookHooks[model] == nil {
		NotebookHooks[model] = map[string]UIHook{}
	}
	NotebookHooks[model][strings.ToLower(pageTitle)] = hook
	// Mirror into swcmeta for SWC workspace rendering.
	swcmeta.RegisterNotebookHook(model, pageTitle, func(ctx context.Context, m string, record map[string]interface{}, ro bool) string {
		vr := &ViewRecordData{ResModel: m, Record: record}
		return string(hook(ctx, vr, ro))
	})
}

type ShellCompanyOption struct {
	ID   int
	Name string
}

// PageData is the top-level template payload for base.html.
type PageData struct {
	Title               string // legacy / diagnostics; prefer ViewBreadcrumb for UI
	ViewBreadcrumb      string // human label for breadcrumb (not the technical model id)
	AppName             string // product display name (browser tab suffix, header)
	ModuleName          string
	Content             template.HTML
	TopMenus            []parser.MenuItem
	SidebarMenus        []SidebarMenu
	ActiveModuleID      string
	ActiveMenuID        string
	ViewStylesheetURLs  []string
	AppsNavActive       bool
	SettingsNavActive   bool
	AppsNavAllowed      bool // base.group_system
	SettingsNavAllowed  bool // base.group_user (internal)
	ExtraStylesheetURLs []string
	LogoURL             string
	// BrandLockupHref is the shell logo/name link target (default: home dashboard via EnrichShellPageData).
	BrandLockupHref string
	// HomeNavHref is the pinned "All apps" top-bar link (default: HomeWebURL).
	HomeNavHref string
	// HomeNavActive highlights the All apps pill on /web/home.
	HomeNavActive bool
	// AppLauncherJSON is installed-app metadata for the global Ctrl+K launcher (JSON array).
	AppLauncherJSON template.JS
	// PinnedAppsJSON is the signed-in user's pinned module list for the shell.
	PinnedAppsJSON     template.JS
	ShellCompany       string
	ShellUser          string
	ShellUserImage     template.URL      // profile photo for top bar (template.URL so data: URLs are not scrubbed); empty → initials
	ShellUserImageCrop template.HTMLAttr // inline crop style for shell avatar when image_crop is set
	UserInitial        string            // legacy single-letter hint; prefer ShellUserInitials in shell chrome
	ShellUserInitials  string            // two-letter avatar label in top bar when no photo
	ShellExtraHTML     template.HTML     // AI Assistant or other shell widgets
	ViewTabs           []ViewSwitchTab   // workspace view switcher in breadcrumb bar; empty hides toolbar

	// ShellCompanyOptions lists companies for the top bar switcher (empty when core.company missing).
	ShellCompanyOptions  []ShellCompanyOption
	ShellActiveCompanyID int // current user's company_id when logged in
	ShowCompanySwitcher  bool
	// UserProfileHref and UserDocsHref are targets for the user dropdown (shell chrome).
	UserProfileHref string
	UserDocsHref    string

	// BreadcrumbTrail: when non-empty, base.html renders linked crumbs; otherwise legacy ModuleName/ViewBreadcrumb.
	BreadcrumbItems []BreadcrumbItem

	// SuppressActivityDock forces the right activity dock off (e.g. Home dashboard) regardless of mail settings.
	SuppressActivityDock bool
	// SuppressSidebar omits the left sidebar entirely (e.g. Home app hub).
	SuppressSidebar bool

	// ExtraBodyClasses is appended to the shell body class list (leading space recommended, e.g. " sum-body--settings-hub").
	ExtraBodyClasses string

	// CSRFToken is injected into POST forms when the user is logged in.
	CSRFToken string

	// FlashMessages are one-time banners (e.g. newly created API key).
	FlashMessages []FlashMessage
	// ToastMessages are success/info notifications shown in the top-right stack.
	ToastMessages []FlashMessage
	// ToastMessagesJSON bootstraps client toasts on first paint.
	ToastMessagesJSON template.JS

	// Right activity panel: Log tab (audit); Messages tab HTML set in RenderView when chatter applies.
	ActivityEnabled         bool
	ActivityLogItems        []ActivityItem
	ActivityContextModel    string
	ActivityContextRecordID int64
	ActivityPanelChatter    bool
	ActivityChatterHTML     template.HTML

	// SWC bootstrap JSON injected as window.__SWC_BOOTSTRAP__
	SWCBootstrapJSON    template.JS
	SWCEnabled          bool
	SwcAddonEntriesJSON template.JS
}

type ActivityItem struct {
	Meta string // author · relative time
	Body string
}

// FlashMessage is a one-time banner in the shell layout.
type FlashMessage struct {
	Kind      string
	Title     string
	Body      string
	Details   string
	ToastOnly bool
}

// ViewSwitchTab is a workspace view mode link in the breadcrumb toolbar.
type ViewSwitchTab struct {
	Label  string
	Href   string
	Mode   string
	Active bool
}

// KanbanColumn is one grouped kanban lane (e.g. a pipeline stage).
type KanbanColumn struct {
	Value    int64
	Label    string
	Sequence int
	Color    int
	Tooltip  string
	Fold     bool
	Records  []map[string]interface{}
}

// SidebarMenu is a sidebar group with child menu links.
type SidebarMenu struct {
	ID       string
	Name     string
	Sequence int
	SubMenus []parser.MenuItem
}

// ViewRecordData carries rows loaded from the ORM for HTML rendering.
type ViewRecordData struct {
	ActionID int
	Record   map[string]interface{}
	ListRows []map[string]interface{}
	ViewTabs []ViewSwitchTab // optional; copied onto PageData for base layout

	// Grouped kanban (when default_group_by is set on the view).
	KanbanColumns    []KanbanColumn
	KanbanGroupField string
	KanbanDraggable  bool
	KanbanModel      string

	// Workspace form chrome (/web): Edit / Save / Cancel and POST save target.
	ResModel      string // e.g. core.company
	RecordID      int    // 0 = create form
	FormEditing   bool   // true when URL contains edit=1
	FormBaseQuery string // query string for /web without leading "?" and without edit= (action, menu_id, view_type, id)
	CSRFToken     string // per-session CSRF hidden field value
	FlashMessages []FlashMessage

	// Pivot aggregation (view type pivot).
	Pivot *PivotData

	// List view quick search (GET q=).
	ListSearchQuery string
	ListSearchURL   string
	ListTotal       int
	ListSort        string
	ListOffset      int
	ListFilter      string
	ListDomain      string
	ListGroupBy     string
}

// PivotData holds aggregated pivot table cells for HTML rendering.
type PivotData struct {
	RowLabels    []string
	ColLabels    []string
	Values       map[string]map[string]float64 // rowKey -> colKey -> sum
	MeasureLabel string
}
