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

// recordBodyText returns a field's body value for storage. Go's `,innerxml` does
// NOT resolve XML entities (e.g. `&amp;` stays literal), so plain character data
// must be unescaped here. Rich content (nested markup such as view arch <xpath>,
// HTML help, etc.) is returned verbatim and re-parsed later by its own consumer.
func recordBodyText(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	// A raw '<' only appears when the body holds nested XML markup; a literal '<'
	// in plain text must be written as &lt; in XML, so it never shows up raw.
	if strings.Contains(trimmed, "<") {
		return trimmed
	}
	return html.UnescapeString(trimmed)
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
		m[f.Name] = recordBodyText(f.Body)
	}
	return m
}
