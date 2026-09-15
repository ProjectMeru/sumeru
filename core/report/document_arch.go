package report

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"strings"
	"text/template"
)

// DocumentRenderContext is the data available to report arch templates.
type DocumentRenderContext struct {
	Record  map[string]interface{}
	User    map[string]interface{}
	Company map[string]interface{}
	Report  map[string]interface{}
}

type documentRoot struct {
	XMLName  xml.Name          `xml:""`
	Title    string            `xml:"title"`
	Subtitle string            `xml:"subtitle"`
	Sections []documentSection `xml:"section"`
	Paras    []string          `xml:"p"`
	Fields   []documentField   `xml:"field"`
	Tables   []documentTable   `xml:"table"`
}

type documentSection struct {
	Heading string          `xml:"heading,attr"`
	Paras   []string        `xml:"p"`
	Fields  []documentField `xml:"field"`
	Tables  []documentTable `xml:"table"`
}

type documentField struct {
	Name string `xml:"name,attr"`
}

type documentTable struct {
	Rows []documentRow `xml:"row"`
}

type documentRow struct {
	Cells []string `xml:"cell"`
}

func parseDocumentArch(arch string) (*documentRoot, error) {
	arch = strings.TrimSpace(arch)
	if arch == "" {
		return nil, fmt.Errorf("report arch is empty")
	}
	var root documentRoot
	if err := xml.Unmarshal([]byte(arch), &root); err != nil {
		return nil, fmt.Errorf("parse report arch: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(root.XMLName.Local)) {
	case "document", "report":
	default:
		return nil, fmt.Errorf("report arch root must be <document> or <report>, got %q", root.XMLName.Local)
	}
	return &root, nil
}

func execTemplateText(raw string, ctx DocumentRenderContext) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if !strings.Contains(raw, "{{") {
		return raw, nil
	}
	tmpl, err := template.New("doc").Option("missingkey=zero").Parse(raw)
	if err != nil {
		return "", fmt.Errorf("report template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("report template exec: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}

func fieldDisplayValue(ctx DocumentRenderContext, name string) string {
	name = strings.TrimSpace(name)
	if name == "" || ctx.Record == nil {
		return ""
	}
	if v, ok := ctx.Record[name]; ok && v != nil {
		return truncatePDFCell(fmt.Sprint(v), 120)
	}
	return ""
}

// TemplatePDFInputFromDocumentArch parses document XML and builds PDF input.
func TemplatePDFInputFromDocumentArch(arch string, ctx DocumentRenderContext, pageSize string) (TemplatePDFInput, error) {
	root, err := parseDocumentArch(arch)
	if err != nil {
		return TemplatePDFInput{}, err
	}
	title, err := execTemplateText(root.Title, ctx)
	if err != nil {
		return TemplatePDFInput{}, err
	}
	subtitle, err := execTemplateText(root.Subtitle, ctx)
	if err != nil {
		return TemplatePDFInput{}, err
	}
	if title == "" {
		title = "Report"
	}
	out := TemplatePDFInput{Title: title, Subtitle: subtitle, PageSize: pageSize}

	for _, p := range root.Paras {
		body, err := execTemplateText(p, ctx)
		if err != nil {
			return TemplatePDFInput{}, err
		}
		if body != "" {
			out.Sections = append(out.Sections, TemplatePDFSection{Body: body})
		}
	}
	for _, f := range root.Fields {
		label := f.Name
		out.TableHead = []string{"Field", "Value"}
		out.TableRows = append(out.TableRows, []string{label, fieldDisplayValue(ctx, f.Name)})
	}
	for _, sec := range root.Sections {
		heading, err := execTemplateText(sec.Heading, ctx)
		if err != nil {
			return TemplatePDFInput{}, err
		}
		var bodies []string
		for _, p := range sec.Paras {
			body, err := execTemplateText(p, ctx)
			if err != nil {
				return TemplatePDFInput{}, err
			}
			if body != "" {
				bodies = append(bodies, body)
			}
		}
		for _, f := range sec.Fields {
			bodies = append(bodies, fmt.Sprintf("%s: %s", f.Name, fieldDisplayValue(ctx, f.Name)))
		}
		out.Sections = append(out.Sections, TemplatePDFSection{
			Heading: heading,
			Body:    strings.Join(bodies, "\n"),
		})
		for _, tbl := range sec.Tables {
			head, rows := tableFromDocument(tbl, ctx)
			if len(head) > 0 {
				out.TableHead = head
				out.TableRows = append(out.TableRows, rows...)
			}
		}
	}
	for _, tbl := range root.Tables {
		head, rows := tableFromDocument(tbl, ctx)
		if len(head) > 0 {
			out.TableHead = head
			out.TableRows = append(out.TableRows, rows...)
		}
	}
	if len(out.Sections) == 0 && len(out.TableRows) == 0 && subtitle == "" {
		return TemplatePDFInput{}, fmt.Errorf("report arch produced no content")
	}
	return out, nil
}

func tableFromDocument(tbl documentTable, ctx DocumentRenderContext) ([]string, [][]string) {
	if len(tbl.Rows) == 0 {
		return nil, nil
	}
	var head []string
	var rows [][]string
	for i, row := range tbl.Rows {
		line := make([]string, len(row.Cells))
		for j, cell := range row.Cells {
			rendered, err := execTemplateText(cell, ctx)
			if err != nil {
				rendered = cell
			}
			line[j] = rendered
		}
		if i == 0 {
			head = line
		} else {
			rows = append(rows, line)
		}
	}
	return head, rows
}

// PreviewHTMLFromDocumentArch renders a simple HTML preview of document arch.
func PreviewHTMLFromDocumentArch(arch string, ctx DocumentRenderContext) (string, error) {
	in, err := TemplatePDFInputFromDocumentArch(arch, ctx, PageSizeA4)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>")
	b.WriteString(html.EscapeString(in.Title))
	b.WriteString("</title></head><body>")
	b.WriteString("<h1>")
	b.WriteString(html.EscapeString(in.Title))
	b.WriteString("</h1>")
	if in.Subtitle != "" {
		b.WriteString("<p><em>")
		b.WriteString(html.EscapeString(in.Subtitle))
		b.WriteString("</em></p>")
	}
	for _, sec := range in.Sections {
		if sec.Heading != "" {
			b.WriteString("<h2>")
			b.WriteString(html.EscapeString(sec.Heading))
			b.WriteString("</h2>")
		}
		if sec.Body != "" {
			b.WriteString("<p>")
			b.WriteString(html.EscapeString(sec.Body))
			b.WriteString("</p>")
		}
	}
	if len(in.TableHead) > 0 {
		b.WriteString("<table border=\"1\"><thead><tr>")
		for _, h := range in.TableHead {
			b.WriteString("<th>")
			b.WriteString(html.EscapeString(h))
			b.WriteString("</th>")
		}
		b.WriteString("</tr></thead><tbody>")
		for _, row := range in.TableRows {
			b.WriteString("<tr>")
			for _, c := range row {
				b.WriteString("<td>")
				b.WriteString(html.EscapeString(c))
				b.WriteString("</td>")
			}
			b.WriteString("</tr>")
		}
		b.WriteString("</tbody></table>")
	}
	b.WriteString("</body></html>")
	return b.String(), nil
}
