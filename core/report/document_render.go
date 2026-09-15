package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/orm"
)

// ReportRenderOptions configures PDF generation and attachment storage.
type ReportRenderOptions struct {
	StoreAttachment bool
}

// LoadReportActionForPreview loads an active report action row (no model ACL check).
func LoadReportActionForPreview(ctx context.Context, reportID int) (map[string]interface{}, error) {
	return loadReportAction(ctx, reportID)
}

// BuildDocumentPDF renders sys.report.action to PDF bytes.
func BuildDocumentPDF(ctx context.Context, actionRow map[string]interface{}, recordID int) ([]byte, string, error) {
	modelName := strings.TrimSpace(orm.AsString(actionRow["model"]))
	title := strings.TrimSpace(orm.AsString(actionRow["name"]))
	pageSize := strings.TrimSpace(orm.AsString(actionRow["paperformat"]))
	if pageSize == "" {
		pageSize = PageSizeA4
	}
	docCtx, err := loadDocumentContext(ctx, actionRow, modelName, recordID)
	if err != nil {
		return nil, "", err
	}
	var pdfIn TemplatePDFInput
	arch := strings.TrimSpace(orm.AsString(actionRow["arch"]))
	templatePath := strings.TrimSpace(orm.AsString(actionRow["template_path"]))
	switch {
	case arch != "":
		pdfIn, err = TemplatePDFInputFromDocumentArch(arch, docCtx, pageSize)
	case templatePath != "":
		pdfIn, err = RenderHTMLTemplate(templatePath, docCtx)
		pdfIn.PageSize = pageSize
	default:
		return nil, "", fmt.Errorf("report action has no arch or template_path")
	}
	if err != nil {
		return nil, "", err
	}
	if pdfIn.Title == "" {
		pdfIn.Title = title
	}
	data, err := ExportTemplatePDF(pdfIn)
	if err != nil {
		return nil, "", err
	}
	safe := strings.ReplaceAll(modelName, ".", "_")
	return data, ExportFilename(safe+"_"+title, "pdf"), nil
}

// RenderReportHTMLPreview returns HTML for admin preview.
func RenderReportHTMLPreview(ctx context.Context, actionRow map[string]interface{}, recordID int) (string, error) {
	modelName := strings.TrimSpace(orm.AsString(actionRow["model"]))
	docCtx, err := loadDocumentContext(ctx, actionRow, modelName, recordID)
	if err != nil {
		return "", err
	}
	arch := strings.TrimSpace(orm.AsString(actionRow["arch"]))
	templatePath := strings.TrimSpace(orm.AsString(actionRow["template_path"]))
	switch {
	case arch != "":
		return PreviewHTMLFromDocumentArch(arch, docCtx)
	case templatePath != "":
		return PreviewHTMLFromTemplate(templatePath, docCtx)
	default:
		return "", fmt.Errorf("report action has no arch or template_path")
	}
}

func loadDocumentContext(ctx context.Context, actionRow map[string]interface{}, modelName string, recordID int) (DocumentRenderContext, error) {
	out := DocumentRenderContext{
		Report: map[string]interface{}{
			"name":          orm.AsString(actionRow["name"]),
			"model":         modelName,
			"template_path": orm.AsString(actionRow["template_path"]),
		},
	}
	if recordID > 0 {
		rec, err := orm.SearchOne(ctx, modelName, map[string]interface{}{"id": recordID})
		if err != nil {
			return DocumentRenderContext{}, err
		}
		out.Record = rec
	} else {
		out.Record = map[string]interface{}{}
	}
	uid := orm.SecurityUID(ctx)
	if uid > 0 {
		if user, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid}); err == nil {
			out.User = user
		}
	}
	if cid := orm.CompanyIDFromContext(ctx); cid > 0 {
		if company, err := orm.SearchOne(ctx, "core.company", map[string]interface{}{"id": cid}); err == nil {
			out.Company = company
		}
	}
	return out, nil
}

// StoreReportPDFAttachment saves PDF to filestore and creates sys.attachment linked to record.
func StoreReportPDFAttachment(ctx context.Context, modelName string, recordID int, filename string, pdf []byte) (int, error) {
	if recordID <= 0 {
		return 0, fmt.Errorf("id required to store attachment")
	}
	if len(pdf) == 0 {
		return 0, fmt.Errorf("empty pdf")
	}
	name := strings.TrimSpace(filename)
	if name == "" {
		name = "report.pdf"
	}
	storeKey := fmt.Sprintf("%s_%d_%s", modelName, recordID, time.Now().UTC().Format("20060102T150405"))
	storeFname, size, err := orm.StoreAttachment(ctx, storeKey, pdf)
	if err != nil {
		return 0, err
	}
	vals := map[string]interface{}{
		"name":        name,
		"model":       modelName,
		"res_id":      recordID,
		"mimetype":    "application/pdf",
		"file_size":   size,
		"store_fname": storeFname,
	}
	if cid := orm.CompanyIDFromContext(ctx); cid > 0 {
		vals["company_id"] = cid
	}
	return orm.Create(ctx, orm.Registry["sys.attachment"], vals)
}
