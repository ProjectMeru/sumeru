package report

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// RenderReportActionPDF loads sys.report.action by id and renders a template PDF.
func RenderReportActionPDF(ctx context.Context, reportID, recordID int) ([]byte, string, error) {
	if reportID <= 0 {
		return nil, "", fmt.Errorf("report id required")
	}
	row, err := orm.SearchOne(ctx, "sys.report.action", map[string]interface{}{"id": reportID})
	if err != nil {
		return nil, "", err
	}
	if !orm.AsBool(row["active"]) {
		return nil, "", fmt.Errorf("report action inactive")
	}
	modelName := strings.TrimSpace(orm.AsString(row["model"]))
	title := strings.TrimSpace(orm.AsString(row["name"]))
	if modelName == "" {
		return nil, "", fmt.Errorf("report model required")
	}
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), modelName, "read"); err != nil {
		return nil, "", err
	}
	pageSize := strings.TrimSpace(orm.AsString(row["paperformat"]))
	if pageSize == "" {
		pageSize = PageSizeA4
	}
	var subtitle string
	if recordID > 0 {
		rec, err := orm.SearchOne(ctx, modelName, map[string]interface{}{"id": recordID})
		if err != nil {
			return nil, "", err
		}
		subtitle = strings.TrimSpace(orm.AsString(rec["name"]))
	}
	data, err := ExportTemplatePDF(TemplatePDFInput{
		Title:    title,
		Subtitle: subtitle,
		Sections: []TemplatePDFSection{
			{Heading: "Report registry", Body: fmt.Sprintf("Model: %s · Template: %s", modelName, orm.AsString(row["template_path"]))},
		},
		TableHead: []string{"Field", "Value"},
		TableRows: [][]string{
			{"Model", modelName},
			{"Record ID", fmt.Sprintf("%d", recordID)},
		},
		PageSize: pageSize,
	})
	if err != nil {
		return nil, "", err
	}
	safe := strings.ReplaceAll(modelName, ".", "_")
	return data, ExportFilename(safe+"_"+title, "pdf"), nil
}

// PivotExportInput configures a pivot read_group CSV export.
type PivotExportInput struct {
	Model       string
	Domain      [][]interface{}
	GroupFields []string
	Measures    []string
}

// GraphExportInput configures a graph read_group CSV export.
type GraphExportInput struct {
	Model        string
	Domain       [][]interface{}
	GroupField   string
	MeasureField string
}

// ExportPivotCSV exports read_group pivot snapshot as CSV.
func ExportPivotCSV(ctx context.Context, in PivotExportInput) ([]byte, error) {
	spec := orm.ReadGroupSpec{GroupBy: in.GroupFields}
	for _, m := range in.Measures {
		spec.Fields = append(spec.Fields, orm.ReadGroupField{Name: m, Field: m, Measure: "sum"})
	}
	rows, err := orm.ReadGroup(ctx, in.Model, in.Domain, spec)
	if err != nil {
		return nil, err
	}
	header := append(append([]string{}, in.GroupFields...), in.Measures...)
	var data [][]string
	for _, row := range rows {
		line := make([]string, len(header))
		for i, g := range in.GroupFields {
			line[i] = fmt.Sprint(row[g])
		}
		for j, m := range in.Measures {
			line[len(in.GroupFields)+j] = fmt.Sprint(row[m])
		}
		data = append(data, line)
	}
	return writeCSV(header, data)
}

// ExportGraphCSV exports read_group graph snapshot as CSV.
func ExportGraphCSV(ctx context.Context, in GraphExportInput) ([]byte, error) {
	return ExportPivotCSV(ctx, PivotExportInput{
		Model:       in.Model,
		Domain:      in.Domain,
		GroupFields: []string{in.GroupField},
		Measures:    []string{in.MeasureField},
	})
}
