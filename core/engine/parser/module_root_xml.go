package parser

import (
	"bytes"
	"encoding/json"
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
	XMLName  xml.Name   `xml:"action"`
	ID       string     `xml:"id,attr"`
	Type     string     `xml:"type,attr"`
	Model    string     `xml:"model,attr"`
	Name     string     `xml:"name,attr"`
	ViewMode     string `xml:"view_mode,attr"`
	ViewID       string `xml:"view_id,attr"`
	SearchViewID string `xml:"search_view_id,attr"`
	Context      string `xml:"context,attr"`
	Domain   string     `xml:"domain,attr"`
	URL      string     `xml:"url,attr"`
	Help     actionHelp `xml:"help"`
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
	ctxMap := map[string]interface{}{}
	if c := strings.TrimSpace(a.Context); c != "" {
		_ = json.Unmarshal([]byte(c), &ctxMap)
	}
	if vid := strings.TrimSpace(a.ViewID); vid != "" {
		ctxMap["view_id"] = vid
	}
	if svid := strings.TrimSpace(a.SearchViewID); svid != "" {
		ctxMap["search_view_id"] = svid
	}
	if len(ctxMap) > 0 {
		if b, err := json.Marshal(ctxMap); err == nil {
			fields = append(fields, RecordField{Name: "context", Body: string(b)})
		}
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
func (v *ViewList) MergeViewListData() {
	if v.Data == nil {
		return
	}
	v.NoUpdate = parseNoUpdateFlag(v.Data.NoUpdate)
	v.Records = append(v.Data.Records, v.Records...)
	v.Views = append(v.Data.Views, v.Views...)
	v.MenuItems = append(v.Data.MenuItems, v.MenuItems...)
	for _, a := range v.Data.Actions {
		v.Records = append(v.Records, a.ToRecord())
	}
	v.Actions = append(v.Data.Actions, v.Actions...)
	v.Data = nil
}

// MergeMenuListData flattens <data> menuitem children.
func (m *MenuList) MergeMenuListData() {
	if m.Data == nil {
		return
	}
	m.NoUpdate = parseNoUpdateFlag(m.Data.NoUpdate)
	m.MenuItems = append(m.Data.MenuItems, m.MenuItems...)
	m.Records = append(m.Data.Records, m.Records...)
	for _, a := range m.Data.Actions {
		m.Records = append(m.Records, a.ToRecord())
	}
	m.Actions = append(m.Data.Actions, m.Actions...)
	m.Data = nil
}

func parseNoUpdateFlag(raw string) bool {
	s := strings.TrimSpace(strings.ToLower(raw))
	return s == "1" || s == "true" || s == "yes"
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
