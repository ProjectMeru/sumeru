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
		m[f.Name] = fieldBodyValue(f)
	}
	return m
}

// fieldBodyValue unwraps CDATA (literal character data, no entity decoding)
// and decodes XML entities for plain text fields. Rich (type="html") bodies
// keep their escaped form so they re-render as valid HTML.
func fieldBodyValue(f RecordField) string {
	body := strings.TrimSpace(f.Body)
	if strings.HasPrefix(body, "<![CDATA[") && strings.HasSuffix(body, "]]>") {
		return strings.TrimSuffix(strings.TrimPrefix(body, "<![CDATA["), "]]>")
	}
	if f.Type == "html" {
		return body
	}
	return html.UnescapeString(body)
}
