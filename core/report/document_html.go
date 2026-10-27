package report

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"
)

var htmlTemplates = map[string]string{}

// RegisterHTMLTemplate registers a named HTML report body (Go html/template).
func RegisterHTMLTemplate(name, body string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	htmlTemplates[name] = body
}

// RenderHTMLTemplate executes a registered HTML template into PDF input.
func RenderHTMLTemplate(name string, ctx DocumentRenderContext) (TemplatePDFInput, error) {
	body, err := executeRegisteredHTML(name, ctx)
	if err != nil {
		return TemplatePDFInput{}, err
	}
	return TemplatePDFInputFromHTML(body, PageSizeA4)
}

func executeRegisteredHTML(name string, ctx DocumentRenderContext) (string, error) {
	raw, ok := htmlTemplates[strings.TrimSpace(name)]
	if !ok {
		return "", fmt.Errorf("report html template %q not found", name)
	}
	tmpl, err := template.New("htmlreport").Option("missingkey=zero").Parse(raw)
	if err != nil {
		return "", fmt.Errorf("html template parse: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("html template exec: %w", err)
	}
	return buf.String(), nil
}

// TemplatePDFInputFromHTML maps a minimal HTML subset to structured PDF sections.
// ponytail: no CSS/layout engine; unknown tags become plain text.
func TemplatePDFInputFromHTML(htmlBody string, pageSize string) (TemplatePDFInput, error) {
	htmlBody = strings.TrimSpace(htmlBody)
	if htmlBody == "" {
		return TemplatePDFInput{}, fmt.Errorf("html report body is empty")
	}
	out := TemplatePDFInput{PageSize: pageSize}
	h1 := regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	h2 := regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`)
	h3 := regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	pRe := regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	if m := h1.FindStringSubmatch(htmlBody); len(m) > 1 {
		out.Title = stripHTMLTags(m[1])
	}
	if out.Title == "" {
		out.Title = "Report"
	}
	for _, m := range h2.FindAllStringSubmatch(htmlBody, -1) {
		out.Sections = append(out.Sections, TemplatePDFSection{Heading: stripHTMLTags(m[1])})
	}
	for _, m := range h3.FindAllStringSubmatch(htmlBody, -1) {
		out.Sections = append(out.Sections, TemplatePDFSection{Heading: stripHTMLTags(m[1])})
	}
	for _, m := range pRe.FindAllStringSubmatch(htmlBody, -1) {
		body := stripHTMLTags(m[1])
		if body == "" {
			continue
		}
		if len(out.Sections) > 0 && out.Sections[len(out.Sections)-1].Body == "" && out.Sections[len(out.Sections)-1].Heading != "" {
			out.Sections[len(out.Sections)-1].Body = body
		} else {
			out.Sections = append(out.Sections, TemplatePDFSection{Body: body})
		}
	}
	tableRe := regexp.MustCompile(`(?is)<table[^>]*>(.*?)</table>`)
	rowRe := regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	cellRe := regexp.MustCompile(`(?is)<t[dh][^>]*>(.*?)</t[dh]>`)
	if tm := tableRe.FindStringSubmatch(htmlBody); len(tm) > 1 {
		var rows [][]string
		for i, rm := range rowRe.FindAllStringSubmatch(tm[1], -1) {
			var cells []string
			for _, cm := range cellRe.FindAllStringSubmatch(rm[1], -1) {
				cells = append(cells, stripHTMLTags(cm[1]))
			}
			if 0 == len(cells) {
				continue
			}
			if 0 == i {
				out.TableHead = cells
			} else {
				rows = append(rows, cells)
			}
		}
		out.TableRows = rows
	}
	if 0 == len(out.Sections) && 0 == len(out.TableRows) {
		out.Sections = append(out.Sections, TemplatePDFSection{Body: stripHTMLTags(htmlBody)})
	}
	return out, nil
}

func stripHTMLTags(s string) string {
	s = regexp.MustCompile(`(?s)<[^>]+>`).ReplaceAllString(s, " ")
	return strings.TrimSpace(html.UnescapeString(s))
}

// PreviewHTMLFromTemplate renders a registered HTML template for admin preview.
func PreviewHTMLFromTemplate(name string, ctx DocumentRenderContext) (string, error) {
	body, err := executeRegisteredHTML(name, ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>%s</title></head><body>%s</body></html>",
		html.EscapeString(name), body), nil
}
