package report

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	gtemplate "github.com/gpdf-dev/gpdf/template"
)

// TemplatePDFSection is one heading + body block in a template-driven PDF.
type TemplatePDFSection struct {
	Heading string
	Body    string
}

// TemplatePDFInput configures a layout PDF (not tabular list export).
type TemplatePDFInput struct {
	Title     string
	Subtitle  string
	Sections  []TemplatePDFSection
	TableHead []string
	TableRows [][]string
	PageSize  string
}

var templatePDFDoc = template.Must(template.New("pdfdoc").Parse(`{{.Title}}
{{if .Subtitle}}{{.Subtitle}}
{{end}}{{range .Sections}}{{.Heading}}
{{.Body}}
{{end}}`))

// ExportTemplatePDF builds a branded document PDF from structured sections (spike for template reports).
func ExportTemplatePDF(in TemplatePDFInput) ([]byte, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "Report"
	}

	pageSize := pdfPageSize(in.PageSize)
	doc := gpdf.NewDocument(
		gpdf.WithPageSize(pageSize),
		gpdf.WithMargins(document.UniformEdges(document.Mm(15))),
		gpdf.WithMetadata(document.DocumentMetadata{Title: title}),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *gtemplate.RowBuilder) {
		r.Col(12, func(c *gtemplate.ColBuilder) {
			c.Text(title, gtemplate.FontSize(18), gtemplate.Bold())
			if sub := strings.TrimSpace(in.Subtitle); sub != "" {
				c.Spacer(document.Mm(2))
				c.Text(sub, gtemplate.FontSize(11), gtemplate.TextColor(pdf.Gray(0.35)))
			}
			c.Spacer(document.Mm(6))
		})
	})

	for _, section := range in.Sections {
		heading := strings.TrimSpace(section.Heading)
		body := strings.TrimSpace(section.Body)
		if heading == "" && body == "" {
			continue
		}
		page.AutoRow(func(r *gtemplate.RowBuilder) {
			r.Col(12, func(c *gtemplate.ColBuilder) {
				if heading != "" {
					c.Text(heading, gtemplate.FontSize(13), gtemplate.Bold())
					c.Spacer(document.Mm(2))
				}
				if body != "" {
					c.Text(body, gtemplate.FontSize(10))
					c.Spacer(document.Mm(4))
				}
			})
		})
	}

	if len(in.TableHead) > 0 {
		rows := in.TableRows
		if rows == nil {
			rows = [][]string{}
		}
		page.AutoRow(func(r *gtemplate.RowBuilder) {
			r.Col(12, func(c *gtemplate.ColBuilder) {
				c.Table(
					in.TableHead,
					rows,
					gtemplate.ColumnWidths(equalColumnWidths(len(in.TableHead))...),
					gtemplate.TableHeaderStyle(gtemplate.Bold(), gtemplate.BgColor(pdf.RGBHex(0xEEEEEE))),
				)
			})
		})
	}

	data, err := doc.Generate()
	if err != nil {
		return nil, fmt.Errorf("template pdf output: %w", err)
	}
	return data, nil
}

// RenderTemplatePDFText renders the plain-text preview of a template PDF (tests / debugging).
func RenderTemplatePDFText(in TemplatePDFInput) (string, error) {
	var buf strings.Builder
	if err := templatePDFDoc.Execute(&buf, in); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
