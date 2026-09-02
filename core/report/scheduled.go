package report

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/mail"
	"sumeru/core/scheduler"
)

// ScheduledReportSpec configures a recurring report email delivery.
type ScheduledReportSpec struct {
	Name       string
	Model      string
	Fields     []string
	Domain     [][]interface{}
	Recipients []string
	Format     string // csv, xlsx, pdf
	CronCode   string
}

// RegisterScheduledReport registers a cron handler that emails a report export.
func RegisterScheduledReport(spec ScheduledReportSpec) error {
	code := strings.TrimSpace(spec.CronCode)
	if code == "" {
		code = "report_scheduled_" + strings.ReplaceAll(strings.ToLower(spec.Name), " ", "_")
	}
	format := strings.ToLower(strings.TrimSpace(spec.Format))
	if format == "" {
		format = "csv"
	}
	model := strings.TrimSpace(spec.Model)
	if model == "" {
		return fmt.Errorf("scheduled report: model required")
	}
	recipients := append([]string(nil), spec.Recipients...)
	fields := append([]string(nil), spec.Fields...)
	domain := spec.Domain
	reportName := spec.Name

	scheduler.RegisterCronHandler(code, func(ctx context.Context, _ map[string]interface{}) error {
		return deliverScheduledReport(ctx, model, fields, domain, format, recipients, reportName)
	})
	return nil
}

func deliverScheduledReport(ctx context.Context, model string, fields []string, domain [][]interface{}, format string, recipients []string, reportName string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("scheduled report %q: no recipients", reportName)
	}
	in := ExportCSVInput{Model: model, Fields: fields, Domain: domain, Title: reportName}
	var (
		data []byte
		err  error
	)
	switch format {
	case "xlsx":
		data, err = ExportXLSX(ctx, in)
	case "pdf":
		data, err = ExportPDF(ctx, ExportPDFInput{Model: model, Fields: fields, Domain: domain, Title: reportName})
	default:
		data, err = ExportCSV(ctx, in)
	}
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("Report: %s", reportName)
	body := fmt.Sprintf("Your scheduled %s export is attached (%d bytes).", format, len(data))
	for _, to := range recipients {
		mail.Enqueue(ctx, strings.TrimSpace(to), subject, body)
	}
	return nil
}
