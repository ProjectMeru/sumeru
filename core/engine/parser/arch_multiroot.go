package parser

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// archFormRoot matches sys.view arch rooted at <form>.
type archFormRoot struct {
	XMLName xml.Name `xml:"form"`
	String  string   `xml:"string,attr"`
	Header  *Header  `xml:"header"`
	Sheet   *Sheet   `xml:"sheet"`
	Footer  *Footer  `xml:"footer"`
	Chatter *Chatter `xml:"chatter"`
	Field   []Field  `xml:"field"`
	Group   []Group  `xml:"group"`
}

type archListRoot struct {
	XMLName         xml.Name       `xml:"list"`
	String          string         `xml:"string,attr"`
	Open            string         `xml:"open,attr"`
	Report          *ReportElement `xml:"report"`
	ReportDownload  string         `xml:"report_download,attr"`
	BulkUpload      string         `xml:"bulk_upload,attr"`
	ReportPDFSizes  string         `xml:"pdf_sizes,attr"`
	ReportBulkModes string         `xml:"bulk_modes,attr"`
	Field           []Field        `xml:"field"`
}

type archKanbanRoot struct {
	XMLName          xml.Name       `xml:"kanban"`
	DefaultGroupBy   string         `xml:"default_group_by,attr"`
	GroupBy          string         `xml:"group_by,attr"`
	RecordsDraggable string         `xml:"records_draggable,attr"`
	QuickCreate      string         `xml:"quick_create,attr"`
	Report           *ReportElement `xml:"report"`
	ReportDownload   string         `xml:"report_download,attr"`
	BulkUpload       string         `xml:"bulk_upload,attr"`
	ReportPDFSizes   string         `xml:"pdf_sizes,attr"`
	ReportBulkModes  string         `xml:"bulk_modes,attr"`
	Field            []Field        `xml:"field"`
}

type archSearchRoot struct {
	XMLName xml.Name       `xml:"search"`
	String  string         `xml:"string,attr"`
	Filter  []SearchFilter `xml:"filter"`
	Field   []Field        `xml:"field"`
}

type archGraphRoot struct {
	XMLName xml.Name `xml:"graph"`
	String  string   `xml:"string,attr"`
	Type    string   `xml:"type,attr"`
	Chart   string   `xml:"chart,attr"`
	Field   []Field  `xml:"field"`
}

type archCalendarRoot struct {
	XMLName   xml.Name `xml:"calendar"`
	String    string   `xml:"string,attr"`
	DateStart string   `xml:"date_start,attr"`
	DateStop  string   `xml:"date_stop,attr"`
	Field     []Field  `xml:"field"`
}

type archPivotRoot struct {
	XMLName xml.Name `xml:"pivot"`
	String  string   `xml:"string,attr"`
	Field   []Field  `xml:"field"`
}

func applyKanbanRootAttrs(v *View, k archKanbanRoot) {
	if v == nil {
		return
	}
	v.DefaultGroupBy = k.DefaultGroupBy
	v.GroupBy = k.GroupBy
	v.RecordsDraggable = k.RecordsDraggable
	v.QuickCreate = k.QuickCreate
}

// promoteNestedForm lifts children of a nested <form> under <view> onto View so
// sheet/header/fields are not lost when XML is <view><form><sheet>…</sheet></form></view>.
func promoteNestedForm(v *View) {
	if v == nil || !formArchHasContent(v.Form) {
		return
	}
	f := v.Form
	if v.Header == nil {
		v.Header = f.Header
	}
	if v.Sheet == nil {
		v.Sheet = f.Sheet
	}
	if v.Footer == nil {
		v.Footer = f.Footer
	}
	if v.Chatter == nil {
		v.Chatter = f.Chatter
	}
	if len(v.Field) == 0 {
		v.Field = f.Field
	}
	if len(v.Group) == 0 {
		v.Group = f.Group
	}
	v.Form = nil
}

func parseViewFromArchInternal(arch string) (*View, error) {
	if arch == "" {
		return nil, fmt.Errorf("empty view arch")
	}

	var v View
	if err := xml.Unmarshal([]byte(arch), &v); err == nil {
		promoteNestedForm(&v)
		if viewLooksPopulated(&v) {
			applyListOpenFlag(&v)
			return &v, nil
		}
	}

	var f archFormRoot
	if err := xml.Unmarshal([]byte(arch), &f); err == nil && formArchHasContent(&f) {
		return &View{
			Type:    "form",
			Header:  f.Header,
			Sheet:   f.Sheet,
			Footer:  f.Footer,
			Chatter: f.Chatter,
			Field:   f.Field,
			Group:   f.Group,
		}, nil
	}

	var l archListRoot
	if err := xml.Unmarshal([]byte(arch), &l); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<list") {
		return &View{
			Type:            "list",
			Field:           l.Field,
			ListOpenAttr:    l.Open,
			ListNoRowOpen:   listOpenAttrDisablesRowNavigation(l.Open),
			Report:          l.Report,
			ReportDownload:  l.ReportDownload,
			BulkUpload:      l.BulkUpload,
			ReportPDFSizes:  l.ReportPDFSizes,
			ReportBulkModes: l.ReportBulkModes,
		}, nil
	}

	var k archKanbanRoot
	if err := xml.Unmarshal([]byte(arch), &k); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<kanban") {
		v := &View{
			Type:            "kanban",
			Field:           k.Field,
			Report:          k.Report,
			ReportDownload:  k.ReportDownload,
			BulkUpload:      k.BulkUpload,
			ReportPDFSizes:  k.ReportPDFSizes,
			ReportBulkModes: k.ReportBulkModes,
		}
		applyKanbanRootAttrs(v, k)
		return v, nil
	}

	var s archSearchRoot
	if err := xml.Unmarshal([]byte(arch), &s); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<search") {
		return &View{
			Type:         "search",
			Title:        s.String,
			Field:        s.Field,
			SearchFilter: s.Filter,
		}, nil
	}

	var g archGraphRoot
	if err := xml.Unmarshal([]byte(arch), &g); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<graph") {
		chart := strings.TrimSpace(g.Chart)
		if chart == "" {
			chart = strings.TrimSpace(g.Type)
		}
		return &View{
			Type:  "graph",
			Title: g.String,
			Chart: chart,
			Field: g.Field,
		}, nil
	}

	var cal archCalendarRoot
	if err := xml.Unmarshal([]byte(arch), &cal); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<calendar") {
		return &View{
			Type:      "calendar",
			Title:     cal.String,
			DateStart: cal.DateStart,
			DateStop:  cal.DateStop,
			Field:     cal.Field,
		}, nil
	}

	var pvt archPivotRoot
	if err := xml.Unmarshal([]byte(arch), &pvt); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<pivot") {
		return &View{
			Type:  "pivot",
			Title: pvt.String,
			Field: pvt.Field,
		}, nil
	}

	if err := xml.Unmarshal([]byte(arch), &v); err != nil {
		return nil, fmt.Errorf("parse view arch: %w", err)
	}
	if !viewLooksPopulated(&v) {
		return nil, fmt.Errorf("parse view arch: unsupported or empty root (use <view>, <form>, <list>, <kanban>, <search>, <graph>, <calendar>, or <pivot>)")
	}
	return &v, nil
}

func viewLooksPopulated(v *View) bool {
	if v == nil {
		return false
	}
	return v.Type != "" || v.Header != nil || v.Sheet != nil || v.Footer != nil || v.Chatter != nil ||
		len(v.Field) > 0 || len(v.Group) > 0 || len(v.SearchFilter) > 0 || v.DateStart != ""
}

func formArchHasContent(f *archFormRoot) bool {
	return f != nil && (f.Header != nil || f.Sheet != nil || f.Footer != nil || f.Chatter != nil ||
		len(f.Field) > 0 || len(f.Group) > 0)
}
