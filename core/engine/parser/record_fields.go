package parser

import (
	"html"
	"strings"
)

// RecordField captures <field> values; use Body (innerxml) for rich content (e.g. arch with xpath).
type RecordField struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Ref  string `xml:"ref,attr"`
	Eval string `xml:"eval,attr"`
	Body string `xml:",innerxml"`
}

// RecordFieldMap returns field name → value. Precedence: ref > eval > body text.
func RecordFieldMap(rec Record) map[string]string {
	m := make(map[string]string)
	for _, f := range rec.Field {
		if f.Name == "" {
			continue
		}
		if f.Ref != "" {
			m[f.Name] = f.Ref
			continue
		}
		if strings.TrimSpace(f.Eval) != "" {
			m[f.Name] = strings.TrimSpace(f.Eval)
			continue
		}
		m[f.Name] = html.UnescapeString(strings.TrimSpace(f.Body))
	}
	return m
}
