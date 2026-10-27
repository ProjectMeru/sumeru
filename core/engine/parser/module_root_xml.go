package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// moduleDataWrapper matches an optional <data noupdate="..."> wrapper inside the module root.
type moduleDataWrapper struct {
	NoUpdate  string     `xml:"noupdate,attr"`
	Records   []Record   `xml:"record"`
	Views     []View     `xml:"view"`
	MenuItems []MenuItem `xml:"menuitem"`
	Actions   []Action   `xml:"action"`
}

// actionHelp captures nested <help>…</help> body (HTML or text).
type actionHelp struct {
	Body string `xml:",innerxml"`
}

// Action represents a simplified <action> tag for actions like sys.action.window.
type Action struct {
	XMLName      xml.Name   `xml:"action"`
	ID           string     `xml:"id,attr"`
	Type         string     `xml:"type,attr"`
	Model        string     `xml:"model,attr"`
	Name         string     `xml:"name,attr"`
	ViewMode     string     `xml:"view_mode,attr"`
	ViewID       string     `xml:"view_id,attr"`
	SearchViewID string     `xml:"search_view_id,attr"`
	Context      string     `xml:"context,attr"`
	Domain       string     `xml:"domain,attr"`
	URL          string     `xml:"url,attr"`
	Target       string     `xml:"target,attr"`
	ResID        string     `xml:"res_id,attr"`
	Limit        string     `xml:"limit,attr"`
	Groups       string     `xml:"groups,attr"`
	Help         actionHelp `xml:"help"`
}

// ToRecord converts an Action to a Record for backward compatibility.
func (a Action) ToRecord() Record {
	if strings.EqualFold(strings.TrimSpace(a.Type), "url") {
		fields := []RecordField{
			{Name: "name", Body: a.Name},
			{Name: "url", Body: a.URL},
		}
		return Record{
			ID:    a.ID,
			Model: "sys.action.url",
			Field: fields,
		}
	}
	fields := []RecordField{
		{Name: "name", Body: a.Name},
		{Name: "core_model", Body: a.Model},
		{Name: "view_mode", Body: a.ViewMode},
	}
	if d := strings.TrimSpace(a.Domain); d != "" {
		fields = append(fields, RecordField{Name: "domain", Body: d})
	}
	if vid := strings.TrimSpace(a.ViewID); vid != "" {
		fields = append(fields, RecordField{Name: "view_id", Body: vid})
	}
	if svid := strings.TrimSpace(a.SearchViewID); svid != "" {
		fields = append(fields, RecordField{Name: "search_view_id", Body: svid})
	}
	if t := strings.TrimSpace(a.Target); t != "" {
		fields = append(fields, RecordField{Name: "target", Body: t})
	}
	if rid := strings.TrimSpace(a.ResID); rid != "" {
		fields = append(fields, RecordField{Name: "res_id", Body: rid})
	}
	if lim := strings.TrimSpace(a.Limit); lim != "" {
		fields = append(fields, RecordField{Name: "limit", Body: lim})
	}
	if g := strings.TrimSpace(a.Groups); g != "" {
		fields = append(fields, RecordField{Name: "groups", Body: g})
	}
	if c := strings.TrimSpace(a.Context); c != "" {
		fields = append(fields, RecordField{Name: "context", Body: c})
	}
	if help := strings.TrimSpace(a.Help.Body); help != "" {
		fields = append(fields, RecordField{Name: "help", Type: "html", Body: help})
	}
	return Record{
		ID:    a.ID,
		Model: "sys.action.window",
		Field: fields,
	}
}

// ValidateModuleRoot returns an error if the module data XML root element is not <sumeru>.
func ValidateModuleRoot(n xml.Name) error {
	switch strings.ToLower(strings.TrimSpace(n.Local)) {
	case "sumeru":
		return nil
	default:
		return fmt.Errorf("module XML root must be <sumeru>, got <%s>", n.Local)
	}
}

// MergeViewListData flattens <data> children into the top-level slices.
func (v *ViewList) MergeViewListData() error {
	if v.Data == nil {
		return nil
	}
	noUpdate, err := ParseXMLBoolAttr("noupdate", v.Data.NoUpdate)
	if err != nil {
		return err
	}
	v.NoUpdate = noUpdate
	v.Records = append(v.Data.Records, v.Records...)
	v.Views = append(v.Data.Views, v.Views...)
	v.MenuItems = append(v.Data.MenuItems, v.MenuItems...)
	for _, a := range v.Data.Actions {
		v.Records = append(v.Records, a.ToRecord())
	}
	v.Actions = append(v.Data.Actions, v.Actions...)
	v.Data = nil
	return nil
}

// MergeMenuListData flattens <data> menuitem children.
func (m *MenuList) MergeMenuListData() error {
	if m.Data == nil {
		return nil
	}
	noUpdate, err := ParseXMLBoolAttr("noupdate", m.Data.NoUpdate)
	if err != nil {
		return err
	}
	m.NoUpdate = noUpdate
	m.MenuItems = append(m.Data.MenuItems, m.MenuItems...)
	m.Records = append(m.Data.Records, m.Records...)
	for _, a := range m.Data.Actions {
		m.Records = append(m.Records, a.ToRecord())
	}
	m.Actions = append(m.Data.Actions, m.Actions...)
	m.Data = nil
	return nil
}

// PeekModuleXMLRootName returns the local name of the first start element (e.g. sumeru).
func PeekModuleXMLRootName(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(bytes.TrimSpace(data)))
	for {
		t, err := dec.Token()
		if err != nil {
			return "", err
		}
		if se, ok := t.(xml.StartElement); ok {
			return se.Name.Local, nil
		}
	}
}
