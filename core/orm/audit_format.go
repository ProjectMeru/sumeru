package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const auditDisplayFieldCap = 8

var auditSkipDiffFields = map[string]bool{
	"id": true, "write_date": true, "write_uid": true,
}

// FormatAuditLogBody builds a human-readable change summary for the activity Log tab (sys.audit only).
func FormatAuditLogBody(ctx context.Context, model, action, beforeJSON, afterJSON, detail string) string {
	action = strings.TrimSpace(action)
	detail = strings.TrimSpace(detail)
	switch action {
	case "create", "insert":
		if body := formatAuditFieldLines(ctx, model, nil, parseAuditJSON(afterJSON), true); body != "" {
			return body
		}
		return "Created record"
	case "unlink", "delete":
		return "Deleted record"
	case "write", "update", "upsert":
		before := parseAuditJSON(beforeJSON)
		after := parseAuditJSON(afterJSON)
		if body := formatAuditDiff(ctx, model, before, after); body != "" {
			return body
		}
	}
	if detail != "" {
		return detail
	}
	if action != "" {
		return action
	}
	return "Change recorded"
}

func parseAuditJSON(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	return m
}

func formatAuditDiff(ctx context.Context, model string, before, after map[string]interface{}) string {
	if before == nil && after == nil {
		return ""
	}
	if before == nil {
		return formatAuditFieldLines(ctx, model, nil, after, false)
	}
	if after == nil {
		after = before
	}
	redact := sensitiveAuditFields(model)
	var keys []string
	seen := map[string]bool{}
	for k := range before {
		if auditSkipDiffFields[k] || redact[k] {
			continue
		}
		seen[k] = true
		if !auditValuesEqual(before[k], after[k]) {
			keys = append(keys, k)
		}
	}
	for k := range after {
		if auditSkipDiffFields[k] || redact[k] || seen[k] {
			continue
		}
		if !auditValuesEqual(before[k], after[k]) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	if 0 == len(keys) {
		return ""
	}
	extra := 0
	if len(keys) > auditDisplayFieldCap {
		extra = len(keys) - auditDisplayFieldCap
		keys = keys[:auditDisplayFieldCap]
	}
	var lines []string
	for _, k := range keys {
		label := auditFieldLabel(ctx, model, k)
		oldV := formatAuditValue(before[k])
		newV := formatAuditValue(after[k])
		if oldV == "" && newV != "" {
			lines = append(lines, fmt.Sprintf("%s: %s", label, newV))
		} else if newV == "" && oldV != "" {
			lines = append(lines, fmt.Sprintf("%s: cleared (was %s)", label, oldV))
		} else {
			lines = append(lines, fmt.Sprintf("%s: %s → %s", label, oldV, newV))
		}
	}
	if extra > 0 {
		lines = append(lines, fmt.Sprintf("and %d more field(s)", extra))
	}
	return strings.Join(lines, "\n")
}

func formatAuditFieldLines(ctx context.Context, model string, _ map[string]interface{}, after map[string]interface{}, createMode bool) string {
	if 0 == len(after) {
		return ""
	}
	redact := sensitiveAuditFields(model)
	var keys []string
	for k, v := range after {
		if auditSkipDiffFields[k] || redact[k] {
			continue
		}
		if formatAuditValue(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if 0 == len(keys) {
		return ""
	}
	extra := 0
	if len(keys) > auditDisplayFieldCap {
		extra = len(keys) - auditDisplayFieldCap
		keys = keys[:auditDisplayFieldCap]
	}
	var lines []string
	for _, k := range keys {
		label := auditFieldLabel(ctx, model, k)
		val := formatAuditValue(after[k])
		if createMode {
			lines = append(lines, fmt.Sprintf("%s: %s", label, val))
		} else {
			lines = append(lines, fmt.Sprintf("%s: %s", label, val))
		}
	}
	if extra > 0 {
		lines = append(lines, fmt.Sprintf("and %d more field(s)", extra))
	}
	return strings.Join(lines, "\n")
}

func sensitiveAuditFields(model string) map[string]bool {
	return readRedactFields(model)
}

func auditFieldLabel(ctx context.Context, model, fieldName string) string {
	defaultLabel := fieldName
	if fd := FieldDef(model, fieldName); fd != nil && strings.TrimSpace(fd.String) != "" {
		defaultLabel = strings.TrimSpace(fd.String)
	}
	return TranslateFieldLabel(ctx, model, fieldName, defaultLabel)
}

func formatAuditValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case bool:
		if x {
			return "Yes"
		}
		return "No"
	case float64:
		if x == float64(int64(x)) {
			return fmt.Sprintf("%d", int64(x))
		}
		return fmt.Sprintf("%v", x)
	case string:
		s := strings.TrimSpace(x)
		if len(s) > 120 {
			return s[:117] + "…"
		}
		return s
	default:
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if len(s) > 120 {
			return s[:117] + "…"
		}
		return s
	}
}

func auditValuesEqual(a, b interface{}) bool {
	return formatAuditValue(a) == formatAuditValue(b)
}
