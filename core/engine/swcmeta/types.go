package swcmeta

// WorkspacePayload is the JSON body for GET /web/swc/workspace.
type WorkspacePayload struct {
	ActionID      int                      `json:"actionId"`
	MenuID        string                   `json:"menuId"`
	ViewType      string                   `json:"viewType"`
	Model         string                   `json:"model"`
	RecordID      int                      `json:"recordId"`
	FormEdit      bool                     `json:"formEdit"`
	CSRFToken     string                   `json:"csrfToken"`
	Arch          ViewArch                 `json:"arch"`
	Record        map[string]interface{}   `json:"record,omitempty"`
	Records       []map[string]interface{} `json:"records,omitempty"`
	ViewTabs      []ViewTab                `json:"viewTabs"`
	Breadcrumbs   []Breadcrumb             `json:"breadcrumbs"`
	ListSearch    string                   `json:"listSearch,omitempty"`
	ListSearchURL string                   `json:"listSearchUrl,omitempty"`
	ListTotal     int                      `json:"listTotal,omitempty"`
	ListSort      string                   `json:"listSort,omitempty"`
	ListOffset    int                      `json:"listOffset,omitempty"`
	ListFilter    string                   `json:"listFilter,omitempty"`
	FormBaseQuery string                   `json:"formBaseQuery,omitempty"`
	Defaults      map[string]interface{}   `json:"defaults,omitempty"`
}

type ViewTab struct {
	Mode   string `json:"mode"`
	Label  string `json:"label"`
	Href   string `json:"href"`
	Active bool   `json:"active"`
}

type Breadcrumb struct {
	Label string `json:"label"`
	Href  string `json:"href,omitempty"`
}

type ViewArch struct {
	Type       string        `json:"type"`
	Model      string        `json:"model"`
	Title      string        `json:"title,omitempty"`
	HasChatter bool          `json:"hasChatter,omitempty"`
	Fields     []ArchField   `json:"fields"`
	Header     *ArchHeader   `json:"header,omitempty"`
	Footer     *ArchFooter   `json:"footer,omitempty"`
	Sheet      *ArchSheet    `json:"sheet,omitempty"`
	FormMeta   *FormMeta     `json:"formMeta,omitempty"`
	Kanban     *KanbanMeta   `json:"kanban,omitempty"`
	Pivot      *PivotMeta    `json:"pivot,omitempty"`
	Search     *SearchMeta   `json:"search,omitempty"`
	Graph      *GraphMeta    `json:"graph,omitempty"`
	Calendar   *CalendarMeta `json:"calendar,omitempty"`
	Report     *ReportMeta   `json:"report,omitempty"`
}

type FormMeta struct {
	HasImageField bool `json:"hasImageField"`
}

type ArchFooter struct {
	Buttons []ArchButton `json:"buttons"`
}

type ArchSeparator struct {
	String string `json:"string,omitempty"`
}

type ArchLabel struct {
	For    string `json:"for,omitempty"`
	String string `json:"string,omitempty"`
}

type ArchField struct {
	Name          string            `json:"name"`
	String        string            `json:"string,omitempty"`
	Type          string            `json:"type,omitempty"`
	Widget        string            `json:"widget,omitempty"`
	Placeholder   string            `json:"placeholder,omitempty"`
	Readonly      bool              `json:"readonly,omitempty"`
	Required      bool              `json:"required,omitempty"`
	Invisible     bool              `json:"invisible,omitempty"`
	ReadonlyExpr  string            `json:"readonly_expr,omitempty"`
	RequiredExpr  string            `json:"required_expr,omitempty"`
	InvisibleExpr string            `json:"invisible_expr,omitempty"`
	PivotType     string            `json:"pivotType,omitempty"`
	Relation      string            `json:"relation,omitempty"`
	Selection     [][]string        `json:"selection,omitempty"`
	Options       map[string]string `json:"options,omitempty"`
	Subview       *ArchListSubview  `json:"subview,omitempty"`
}

type ArchListSubview struct {
	Editable string      `json:"editable,omitempty"`
	Fields   []ArchField `json:"fields"`
}

type ArchButton struct {
	Name   string `json:"name"`
	String string `json:"string"`
	Type   string `json:"type"`
	Class  string `json:"class,omitempty"`
}

type ArchHeader struct {
	Buttons []ArchButton `json:"buttons"`
	Fields  []ArchField  `json:"fields"`
}

type ArchSheet struct {
	Divs       []ArchDiv       `json:"divs,omitempty"`
	Groups     []ArchGroup     `json:"groups"`
	Notebook   []ArchNotebook  `json:"notebook,omitempty"`
	Fields     []ArchField     `json:"fields"`
	Separators []ArchSeparator `json:"separators,omitempty"`
	Labels     []ArchLabel     `json:"labels,omitempty"`
}

type ArchDiv struct {
	Class    string       `json:"class,omitempty"`
	Fields   []ArchField  `json:"fields,omitempty"`
	Buttons  []ArchButton `json:"buttons,omitempty"`
	H1Fields []ArchField  `json:"h1Fields,omitempty"`
	Divs     []ArchDiv    `json:"divs,omitempty"`
}

type ArchGroup struct {
	String     string          `json:"string,omitempty"`
	Col        int             `json:"col,omitempty"`
	Colspan    int             `json:"colspan,omitempty"`
	Fields     []ArchField     `json:"fields"`
	Groups     []ArchGroup     `json:"groups,omitempty"`
	Separators []ArchSeparator `json:"separators,omitempty"`
	Labels     []ArchLabel     `json:"labels,omitempty"`
}

type ArchNotebook struct {
	Pages []ArchPage `json:"pages"`
}

type ArchPage struct {
	Title      string          `json:"title"`
	Groups     []ArchGroup     `json:"groups"`
	Fields     []ArchField     `json:"fields"`
	Separators []ArchSeparator `json:"separators,omitempty"`
	Labels     []ArchLabel     `json:"labels,omitempty"`
}

type KanbanMeta struct {
	GroupField  string         `json:"groupField"`
	Draggable   bool           `json:"draggable"`
	QuickCreate bool           `json:"quickCreate,omitempty"`
	Columns     []KanbanColumn `json:"columns"`
}

type KanbanColumn struct {
	Value    int64                    `json:"value"`
	Label    string                   `json:"label"`
	Sequence int                      `json:"sequence"`
	Color    int                      `json:"color"`
	Fold     bool                     `json:"fold"`
	Records  []map[string]interface{} `json:"records"`
}

type PivotMeta struct {
	RowLabels    []string                      `json:"rowLabels"`
	ColLabels    []string                      `json:"colLabels"`
	Values       map[string]map[string]float64 `json:"values"`
	MeasureLabel string                        `json:"measureLabel"`
}

type SearchFilterMeta struct {
	Name    string `json:"name"`
	String  string `json:"string"`
	Domain  string `json:"domain,omitempty"`
	GroupBy string `json:"groupBy,omitempty"`
}

type SearchMeta struct {
	Filters []SearchFilterMeta `json:"filters"`
}

type GraphMeta struct {
	Chart string `json:"chart,omitempty"`
}

type CalendarMeta struct {
	DateStart string `json:"dateStart,omitempty"`
	DateStop  string `json:"dateStop,omitempty"`
}

type ReportMeta struct {
	Download  bool   `json:"download"`
	Upload    bool   `json:"upload"`
	PDFSizes  string `json:"pdfSizes,omitempty"`
	BulkModes string `json:"bulkModes,omitempty"`
}
