package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

type View struct {
	XMLName  xml.Name `xml:"view"`
	ID       string   `xml:"id,attr"`
	Model    string   `xml:"model,attr"`
	Type     string   `xml:"type,attr"`
	Title    string   `xml:"title,attr"`
	Priority int      `xml:"priority,attr"`
	// ListOpenAttr is the raw <list open="..."/> or <view type="list" open="..."/> attribute (false/0/off disables row→form).
	ListOpenAttr string `xml:"open,attr"`
	// ListNoRowOpen is derived from ListOpenAttr by the arch parser for type list.
	ListNoRowOpen bool     `xml:"-"`
	Header        *Header  `xml:"header"`
	Sheet         *Sheet   `xml:"sheet"`
	Footer        *Footer  `xml:"footer"`
	Chatter       *Chatter `xml:"chatter"`
	Field         []Field  `xml:"field"`
	Group         []Group  `xml:"group"`
	// Form is a nested <form> under <view>; promoted onto View then cleared.
	Form *archFormRoot `xml:"form"`

	// Kanban grouping / drag-drop (view type="kanban" or nested kanban arch).
	DefaultGroupBy   string `xml:"default_group_by,attr"`
	GroupBy          string `xml:"group_by,attr"`
	RecordsDraggable string `xml:"records_draggable,attr"`
	QuickCreate      string `xml:"quick_create,attr"`

	// Calendar / graph extras (ignored on other view types).
	Chart     string `xml:"chart,attr"`
	DateStart string `xml:"date_start,attr"`
	DateStop  string `xml:"date_stop,attr"`

	// Search view filters (<filter name domain group_by>).
	SearchFilter []SearchFilter `xml:"filter"`

	// Report exchange (download CSV/PDF, bulk upload) — opt-in per view.
	Report          *ReportElement `xml:"report"`
	ReportDownload  string         `xml:"report_download,attr"`
	BulkUpload      string         `xml:"bulk_upload,attr"`
	ReportPDFSizes  string         `xml:"pdf_sizes,attr"`
	ReportBulkModes string         `xml:"bulk_modes,attr"`
}

// ReportElement declares report download and bulk upload on a view.
type ReportElement struct {
	Download string `xml:"download,attr"`
	Upload   string `xml:"upload,attr"`
	PDFSizes string `xml:"pdf_sizes,attr"`
	Modes    string `xml:"modes,attr"`
}

// KanbanGroupField returns the column grouping field (default_group_by, then group_by).
func (v *View) KanbanGroupField() string {
	if v == nil {
		return ""
	}
	if g := strings.TrimSpace(v.DefaultGroupBy); g != "" {
		return g
	}
	return strings.TrimSpace(v.GroupBy)
}

// KanbanDraggable is true when records_draggable is not explicitly "0"/"false".
func (v *View) KanbanDraggable() bool {
	if v == nil {
		return false
	}
	if v.KanbanGroupField() == "" {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(v.RecordsDraggable))
	if s == "0" || s == "false" || s == "off" || s == "no" {
		return false
	}
	return true
}

// KanbanQuickCreate is true when grouped and quick_create is not explicitly off.
func (v *View) KanbanQuickCreate() bool {
	if v == nil || v.KanbanGroupField() == "" {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(v.QuickCreate))
	if s == "0" || s == "false" || s == "off" || s == "no" {
		return false
	}
	return true
}

// GraphChart returns the chart kind (bar, line, pie); default bar.
func (v *View) GraphChart() string {
	if v == nil {
		return "bar"
	}
	c := strings.ToLower(strings.TrimSpace(v.Chart))
	if c == "" {
		return "bar"
	}
	return c
}

type Header struct {
	Button []Button `xml:"button"`
	Field  []Field  `xml:"field"`
	Widget []Widget `xml:"widget"`
}

// Widget is a header toolbar widget (e.g. report_download, bulk_upload).
type Widget struct {
	Type     string `xml:"type,attr"`
	Formats  string `xml:"formats,attr"`
	PDFSizes string `xml:"pdf_sizes,attr"`
	Modes    string `xml:"modes,attr"`
}

type Sheet struct {
	Div       []Div       `xml:"div"`
	Group     []Group     `xml:"group"`
	Notebook  []Notebook  `xml:"notebook"`
	Field     []Field     `xml:"field"`
	Separator []Separator `xml:"separator"`
	Label     []Label     `xml:"label"`
}

type Notebook struct {
	Page []Page `xml:"page"`
}

type Page struct {
	Title     string      `xml:"string,attr"`
	Field     []Field     `xml:"field"`
	Group     []Group     `xml:"group"`
	Separator []Separator `xml:"separator"`
	Label     []Label     `xml:"label"`
}

type Button struct {
	Name   string `xml:"name,attr"`
	String string `xml:"string,attr"`
	Type   string `xml:"type,attr"`
	Class  string `xml:"class,attr"`
}

type Chatter struct {
	Field []Field `xml:"field"`
}

type Div struct {
	Class  string   `xml:"class,attr"`
	Field  []Field  `xml:"field"`
	Button []Button `xml:"button"`
	H1     []H1     `xml:"h1"`
	Div    []Div    `xml:"div"`
}

type H1 struct {
	Field []Field `xml:"field"`
}

type Field struct {
	Name        string     `xml:"name,attr"`
	Label       string     `xml:"string,attr"`
	Widget      string     `xml:"widget,attr"`
	Placeholder string     `xml:"placeholder,attr"`
	Options     string     `xml:"options,attr"`
	Groups      string     `xml:"groups,attr"`
	Invisible   string     `xml:"invisible,attr"`
	Readonly    string     `xml:"readonly,attr"`
	Required    string     `xml:"required,attr"`
	PivotType   string     `xml:"type,attr"` // row | col | measure (pivot views only)
	List        *FieldList `xml:"list"`
	Tree        *FieldList `xml:"tree"`
}

// FieldList is an embedded list/tree subview on a form field (O2M/M2M).
type FieldList struct {
	Editable string  `xml:"editable,attr"`
	Field    []Field `xml:"field"`
}

// SearchFilter is a named search-view facet (domain JSON and optional group_by).
type SearchFilter struct {
	Name    string `xml:"name,attr"`
	String  string `xml:"string,attr"`
	Domain  string `xml:"domain,attr"`
	GroupBy string `xml:"group_by,attr"`
}

type Group struct {
	Title     string      `xml:"string,attr"`
	Col       string      `xml:"col,attr"`
	Colspan   string      `xml:"colspan,attr"`
	Field     []Field     `xml:"field"`
	Group     []Group     `xml:"group"`
	Separator []Separator `xml:"separator"`
	Label     []Label     `xml:"label"`
}

// Footer is the form footer (action buttons).
type Footer struct {
	Button []Button `xml:"button"`
}

// Separator is a visual section break (<separator/>).
type Separator struct {
	String string `xml:"string,attr"`
}

// Label is a static label optionally tied to a field (<label for="..."/>).
type Label struct {
	For    string `xml:"for,attr"`
	String string `xml:"string,attr"`
}

type Menu struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	ParentID string `xml:"parent,attr"`
	Action   string `xml:"action,attr"`
	Sequence int    `xml:"sequence,attr"`
}

type Record struct {
	ID    string        `xml:"id,attr"`
	Model string        `xml:"model,attr"`
	Field []RecordField `xml:"field"`
}

type MenuItem struct {
	ID           string `xml:"id,attr"`
	Name         string `xml:"name,attr"`
	ParentID     string `xml:"parent,attr"`
	Action       string `xml:"action,attr"`
	Sequence     int    `xml:"sequence,attr"`
	WebIcon      string `xml:"web_icon,attr"`
	AccessGroups string `xml:"groups,attr"`
	Module       string `xml:"-"` // set from DB rows only (not in XML menu files)
}

type ViewList struct {
	XMLName   xml.Name           `xml:"sumeru"`
	NoUpdate  bool               `xml:"-"`
	Data      *moduleDataWrapper `xml:"data"`
	Records   []Record           `xml:"record"`
	Views     []View             `xml:"view"`
	MenuItems []MenuItem         `xml:"menuitem"`
	Actions   []Action           `xml:"action"`
}

type MenuList struct {
	XMLName   xml.Name           `xml:"sumeru"`
	NoUpdate  bool               `xml:"-"`
	Data      *moduleDataWrapper `xml:"data"`
	MenuItems []MenuItem         `xml:"menuitem"`
	Records   []Record           `xml:"record"`
	Actions   []Action           `xml:"action"`
}

func ParseViewList(filePath string) (*ViewList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	root, err := PeekModuleXMLRootName(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filePath, err)
	}
	if err := ValidateModuleRoot(xml.Name{Local: root}); err != nil {
		return nil, fmt.Errorf("%s: %w", filePath, err)
	}
	var viewList ViewList
	if err := xml.Unmarshal(data, &viewList); err != nil {
		return nil, err
	}
	viewList.MergeViewListData()
	for i := range viewList.Views {
		promoteNestedForm(&viewList.Views[i])
	}
	return &viewList, nil
}

func ParseViewFromArch(arch string) (*View, error) {
	v, err := parseViewFromArchInternal(strings.TrimSpace(arch))
	if err != nil {
		return nil, err
	}
	v.Type = strings.ToLower(strings.TrimSpace(v.Type))
	return v, nil
}

func ParseMenuList(filePath string) (*MenuList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	root, err := PeekModuleXMLRootName(data)
	if err != nil {
		return nil, err
	}
	if err := ValidateModuleRoot(xml.Name{Local: root}); err != nil {
		return nil, err
	}
	var menuList MenuList
	if err := xml.Unmarshal(data, &menuList); err != nil {
		return nil, err
	}
	menuList.MergeMenuListData()
	return &menuList, nil
}
