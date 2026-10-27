package audit

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"sumeru/core/sdk"
)

const (
	auditArchiveStateStored = "stored"
	auditArchiveStateFailed = "failed"
	auditArchiveStatePurged = "purged"
)

// auditExportRow is one sys.audit row in NDJSON export.
type auditExportRow struct {
	ID         int64  `json:"id"`
	Action     string `json:"action"`
	Model      string `json:"model"`
	ResID      int64  `json:"res_id"`
	UserID     int64  `json:"user_id,omitempty"`
	CreateDate string `json:"create_date"`
	BeforeJSON string `json:"before_json,omitempty"`
	AfterJSON  string `json:"after_json,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

func auditRowFromMap(row map[string]interface{}) auditExportRow {
	id, _ := sdk.CoerceInt64(row["id"])
	resID, _ := sdk.CoerceInt64(row["res_id"])
	uid, _ := sdk.CoerceInt64(row["user_id"])
	return auditExportRow{
		ID:         id,
		Action:     sdk.AsString(row["action"]),
		Model:      sdk.AsString(row["model"]),
		ResID:      resID,
		UserID:     uid,
		CreateDate: sdk.AsString(row["create_date"]),
		BeforeJSON: sdk.AsString(row["before_json"]),
		AfterJSON:  sdk.AsString(row["after_json"]),
		Detail:     sdk.AsString(row["detail"]),
	}
}

func encodeAuditJSONL(rows []auditExportRow) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return nil, fmt.Errorf("jsonl encode: %w", err)
		}
	}
	return buf.Bytes(), nil
}

func gzipBytes(raw []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(raw); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func maybeGzipAudit(compress bool, raw []byte) (payload []byte, mime string, err error) {
	if !compress {
		return raw, "application/x-ndjson", nil
	}
	out, err := gzipBytes(raw)
	if err != nil {
		return nil, "", err
	}
	return out, "application/gzip", nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func auditPeriodBounds(rows []auditExportRow) (from, to string) {
	if 0 == len(rows) {
		return "", ""
	}
	dates := make([]string, len(rows))
	for i, r := range rows {
		dates[i] = r.CreateDate
	}
	sort.Strings(dates)
	return dates[0], dates[len(dates)-1]
}

func auditMinMaxID(rows []auditExportRow) (minID, maxID int64) {
	if 0 == len(rows) {
		return 0, 0
	}
	minID = rows[0].ID
	maxID = rows[0].ID
	for _, r := range rows[1:] {
		if r.ID < minID {
			minID = r.ID
		}
		if r.ID > maxID {
			maxID = r.ID
		}
	}
	return minID, maxID
}

// DecodeAuditJSONLForTest parses NDJSON (tests).
func DecodeAuditJSONLForTest(data []byte) ([]auditExportRow, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var rows []auditExportRow
	for {
		var row auditExportRow
		if err := dec.Decode(&row); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// EncodeAuditJSONLForTest encodes rows (tests).
func EncodeAuditJSONLForTest(rows []auditExportRow) ([]byte, error) {
	return encodeAuditJSONL(rows)
}

// MaybeGzipAuditForTest compresses export payload (tests).
func MaybeGzipAuditForTest(compress bool, raw []byte) ([]byte, string, error) {
	return maybeGzipAudit(compress, raw)
}

// SHA256HexAuditForTest returns digest hex (tests).
func SHA256HexAuditForTest(data []byte) string {
	return sha256Hex(data)
}

// AuditPeriodBoundsForTest returns min/max create_date (tests).
func AuditPeriodBoundsForTest(rows []auditExportRow) (string, string) {
	return auditPeriodBounds(rows)
}

// AuditMinMaxIDForTest returns min/max row id (tests).
func AuditMinMaxIDForTest(rows []auditExportRow) (int64, int64) {
	return auditMinMaxID(rows)
}

// AuditRowFromMapForTest maps ORM row to export row (tests).
func AuditRowFromMapForTest(row map[string]interface{}) AuditExportRowForTest {
	return auditRowFromMap(row)
}
