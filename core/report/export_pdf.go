package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

// ExportPDFInput configures a PDF export.
type ExportPDFInput struct {
	Model    string
	Fields   []string
	Domain   [][]interface{}
	RecordID int
	Title    string
	PageSize string
}

func pdfPageSize(name string) document.Size {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case PageSizeLegal:
		return gpdf.Legal
	case PageSizeLetter:
		return gpdf.Letter
	default:
		return gpdf.A4
	}
}

func landscapePage(size document.Size) document.Size {
	return document.Size{Width: size.Height, Height: size.Width}
}

func equalColumnWidths(n int) []float64 {
	if 0 >= n {
		return nil
	}
	width := 100.0 / float64(n)
	out := make([]float64, n)
	for i := range out {
		out[i] = width
	}
	return out
}

func truncatePDFCell(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max-3] + "..."
}

// ExportPDF builds a tabular PDF for the given model rows and fields.
func ExportPDF(ctx context.Context, in ExportPDFInput) ([]byte, error) {
	fields, err := ValidateFields(in.Model, in.Fields)
	if err != nil {
		return nil, err
	}
	rows, err := FetchRows(ctx, in.Model, in.Domain, in.RecordID)
	if err != nil {
		return nil, err
	}

	pageSize := pdfPageSize(in.PageSize)
	if len(fields) > 6 {
		pageSize = landscapePage(pageSize)
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = in.Model
	}

	labels := FieldLabels(in.Model, fields)
	header := make([]string, len(fields))
	for i, f := range fields {
		if lbl := labels[f]; lbl != "" {
			header[i] = lbl
		} else {
			header[i] = f
		}
	}

	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		line := make([]string, len(fields))
		for i, f := range fields {
			line[i] = truncatePDFCell(formatCell(ctx, in.Model, f, row[f]), 80)
		}
		tableRows = append(tableRows, line)
	}

	cellBorder := template.Border(
		template.BorderWidth(document.Pt(0.5)),
		template.BorderColor(pdf.Gray(0.75)),
	)

	doc := gpdf.NewDocument(
		gpdf.WithPageSize(pageSize),
		gpdf.WithMargins(document.UniformEdges(document.Mm(15))),
		gpdf.WithMetadata(document.DocumentMetadata{Title: title}),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text(title, template.FontSize(16), template.Bold())
			c.Spacer(document.Mm(4))
		})
	})
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				header,
				tableRows,
				template.ColumnWidths(equalColumnWidths(len(fields))...),
				template.TableHeaderStyle(template.Bold(), template.BgColor(pdf.RGBHex(0xEEEEEE))),
				template.WithTableCellBorder(cellBorder),
			)
		})
	})

	data, err := doc.Generate()
	if err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return data, nil
}
