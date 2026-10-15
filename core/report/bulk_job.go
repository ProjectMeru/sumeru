package report

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/scheduler"
)

func init() {
	scheduler.RegisterCronHandler(BulkJobCronCode, runBulkJobCron)
}

// RunBulkJobCronForTest executes one queued bulk job (tests only).
func RunBulkJobCronForTest(ctx context.Context) error {
	return runBulkJobCron(ctx, nil)
}

func runBulkJobCron(ctx context.Context, _ map[string]interface{}) error {
	batch, err := orm.SearchOne(ctx, BulkModelName, map[string]interface{}{"state": BulkStateQueued})
	if err != nil {
		return nil
	}
	if 0 == len(batch) {
		return nil
	}
	id, _ := orm.CoerceInt64(batch["id"])
	if 0 >= id {
		return nil
	}
	queuedUserID, _ := orm.CoerceInt64(batch["user_id"])
	if 0 >= queuedUserID {
		_ = orm.UpdateRecordByID(ctx, BulkModelName, int(id), map[string]interface{}{
			"state":          BulkStateFailed,
			"result_summary": "Missing user for queued job",
		})
		return nil
	}
	_ = orm.UpdateRecordByID(ctx, BulkModelName, int(id), map[string]interface{}{"state": BulkStateRunning})
	direction := strings.TrimSpace(orm.AsString(batch["direction"]))
	if direction == DirectionExport {
		err = runExportJob(ctx, batch)
	} else {
		mapping := map[string]string{}
		_ = json.Unmarshal([]byte(orm.AsString(batch["column_mapping"])), &mapping)
		_, err = ExecuteBulkImport(ctx, ExecuteBulkImportInput{
			BatchID:     int(id),
			Mapping:     mapping,
			SkipInvalid: true,
		})
	}
	if err != nil {
		_ = orm.UpdateRecordByID(ctx, BulkModelName, int(id), map[string]interface{}{
			"state":          BulkStateFailed,
			"result_summary": err.Error(),
		})
		return err
	}
	if direction != DirectionExport {
		_ = orm.UpdateRecordByID(ctx, BulkModelName, int(id), map[string]interface{}{
			"state":          BulkStateDone,
			"result_summary": "Import finished",
		})
	}
	return nil
}

// RunExportJobForTest runs export job logic for external tests.
func RunExportJobForTest(ctx context.Context, batch map[string]interface{}) error {
	return runExportJob(ctx, batch)
}

func runExportJob(ctx context.Context, batch map[string]interface{}) error {
	modelName := strings.TrimSpace(orm.AsString(batch["target_model"]))
	format := strings.TrimSpace(orm.AsString(batch["export_format"]))
	if format == "" {
		format = "csv"
	}
	var domain [][]interface{}
	_ = json.Unmarshal([]byte(orm.AsString(batch["export_domain"])), &domain)
	fields := ParseFieldsParam(orm.AsString(batch["selected_fields"]))
	in := ExportCSVInput{Model: modelName, Fields: fields, Domain: domain}
	queuedUserCtx := bulkActorContext(batch)
	var data []byte
	var err error
	data, err = ExportCSVUnlimited(queuedUserCtx, in)
	if format == "xlsx" {
		fields, _ := ValidateFields(modelName, fields)
		rows, _ := FetchRowsUnlimited(queuedUserCtx, modelName, domain, 0)
		var table [][]string
		for _, row := range rows {
			line := make([]string, len(fields))
			for i, f := range fields {
				line[i] = formatCell(queuedUserCtx, modelName, f, row[f])
			}
			table = append(table, line)
		}
		data, err = ExportXLSXForTest(fields, table)
	}
	if err != nil {
		return err
	}
	batchID, _ := orm.CoerceInt64(batch["id"])
	name := fmt.Sprintf("export_%s.%s", modelName, format)
	attID, err := orm.CreateBinaryAttachment(ctx, orm.CreateBinaryAttachmentInput{
		Name:     name,
		ResModel: BulkModelName,
		ResID:    int(batchID),
		Data:     data,
		Mimetype: exportMime(format),
	})
	if err != nil {
		return err
	}
	return orm.UpdateRecordByID(ctx, BulkModelName, int(batchID), map[string]interface{}{
		"state":          BulkStateDone,
		"attachment_id":  attID,
		"result_summary": fmt.Sprintf("Export ready (%s)", name),
	})
}

// bulkActorContext is the queued user, without the scheduler bypass flag.
func bulkActorContext(batch map[string]interface{}) context.Context {
	uid64, _ := orm.CoerceInt64(batch["user_id"])
	ctx := orm.ContextWithUID(context.Background(), int(uid64))
	if cid := orm.ActiveCompanyIDForUser(ctx, int(uid64)); cid > 0 {
		ctx = orm.ContextWithCompanyID(ctx, cid)
	}
	return ctx
}

// BulkJobActorContextForTest exposes the queued-user context for external tests.
func BulkJobActorContextForTest(batch map[string]interface{}) context.Context {
	return bulkActorContext(batch)
}

func exportMime(format string) string {
	if format == "xlsx" {
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	return "text/csv"
}

// QueueExportJob creates an async export batch.
func QueueExportJob(ctx context.Context, modelName string, fields []string, domain [][]interface{}, format string, uid int) (int, error) {
	fieldsJSON, _ := json.Marshal(fields)
	domainJSON, _ := json.Marshal(domain)
	inst, ok := orm.Registry[BulkModelName]
	if !ok {
		return 0, fmt.Errorf("bulk import model missing")
	}
	return orm.Create(ctx, inst, map[string]interface{}{
		"name":            fmt.Sprintf("Export %s", modelName),
		"target_model":    modelName,
		"direction":       DirectionExport,
		"export_format":   format,
		"export_domain":   string(domainJSON),
		"selected_fields": string(fieldsJSON),
		"user_id":         uid,
		"state":           BulkStateQueued,
	})
}

// CountBatchRows returns CSV row count for a batch.
func CountBatchRows(ctx context.Context, batchID int) (int, error) {
	_, data, _, err := loadBatchCSV(ctx, batchID)
	if err != nil {
		return 0, err
	}
	_, rows, err := parseCSV(data)
	return len(rows), err
}
