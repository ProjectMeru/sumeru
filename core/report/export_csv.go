package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/orm"
)

// ExportCSVInput configures a CSV export.
type ExportCSVInput struct {
	Model    string
	Fields   []string
	Domain   [][]interface{}
	RecordID int
	Title    string
}

// ExportCSV builds CSV bytes for the given model rows and fields.
func ExportCSV(ctx context.Context, in ExportCSVInput) ([]byte, error) {
	fields, err := ValidateFields(in.Model, in.Fields)
	if err != nil {
		return nil, err
	}
	if err := checkExportFieldAccess(ctx, in.Model, fields); err != nil {
		return nil, err
	}
	rows, err := FetchRows(ctx, in.Model, in.Domain, in.RecordID)
	if err != nil {
		return nil, err
	}
	var data [][]string
	for _, row := range rows {
		line := make([]string, len(fields))
		for i, f := range fields {
			line[i] = formatCell(ctx, in.Model, f, row[f])
		}
		data = append(data, line)
	}
	return writeCSV(fields, data)
}

// ExportCSVUnlimitedForTest exposes unlimited CSV export for external tests.
func ExportCSVUnlimitedForTest(ctx context.Context, in ExportCSVInput) ([]byte, error) {
	return ExportCSVUnlimited(ctx, in)
}

// ExportCSVUnlimited exports all matching rows (no 500 cap).
func ExportCSVUnlimited(ctx context.Context, in ExportCSVInput) ([]byte, error) {
	fields, err := ValidateFields(in.Model, in.Fields)
	if err != nil {
		return nil, err
	}
	if err := checkExportFieldAccess(ctx, in.Model, fields); err != nil {
		return nil, err
	}
	rows, err := FetchRowsUnlimited(ctx, in.Model, in.Domain, in.RecordID)
	if err != nil {
		return nil, err
	}
	var data [][]string
	for _, row := range rows {
		line := make([]string, len(fields))
		for i, f := range fields {
			line[i] = formatCell(ctx, in.Model, f, row[f])
		}
		data = append(data, line)
	}
	return writeCSV(fields, data)
}

// BulkTemplateCSV returns header-only CSV for selected fields.
func BulkTemplateCSV(modelName string, fields []string) ([]byte, error) {
	fields, err := ValidateFields(modelName, fields)
	if err != nil {
		return nil, err
	}
	return writeCSV(fields, nil)
}

// ExportFilename builds a download filename.
func ExportFilename(modelName, ext string) string {
	safe := strings.ReplaceAll(modelName, ".", "_")
	return fmt.Sprintf("%s_%s.%s", safe, time.Now().Format("20060102_150405"), ext)
}

// DefaultFieldsFromView returns field names from a parsed view arch field list.
func DefaultFieldsFromView(fieldNames []string) []string {
	var out []string
	for _, n := range fieldNames {
		n = strings.TrimSpace(n)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

// FieldLabels returns display labels for fields from model metadata.
func FieldLabels(modelName string, fields []string) map[string]string {
	labels := map[string]string{}
	if m, ok := orm.Registry[modelName]; ok {
		for _, f := range m.Fields() {
			if f.String != "" {
				labels[f.Name] = f.String
			} else {
				labels[f.Name] = f.Name
			}
		}
	}
	for _, f := range fields {
		if _, ok := labels[f]; !ok {
			labels[f] = f
		}
	}
	return labels
}
