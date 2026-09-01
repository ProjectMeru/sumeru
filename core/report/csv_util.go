package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// sanitizeSpreadsheetCell prefixes values that spreadsheet apps may interpret as formulas.
func sanitizeSpreadsheetCell(value string) string {
	if value == "" {
		return value
	}
	trimmed := strings.TrimSpace(value)
	switch value[0] {
	case '=':
		return "\t" + value
	case '+', '-':
		if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return value
		}
		return "\t" + value
	case '@', '\t', '\r':
		return "\t" + value
	}
	return value
}

// SanitizeSpreadsheetCellForTest exposes spreadsheet injection guard for tests.
func SanitizeSpreadsheetCellForTest(value string) string {
	return sanitizeSpreadsheetCell(value)
}

func writeCSV(headers []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	for _, row := range rows {
		safeRow := make([]string, len(row))
		for i, cell := range row {
			safeRow[i] = sanitizeSpreadsheetCell(cell)
		}
		if err := w.Write(safeRow); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func parseCSV(content []byte) (headers []string, rows [][]string, err error) {
	r := csv.NewReader(bytes.NewReader(content))
	r.TrimLeadingSpace = true
	headers, err = r.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("empty csv")
	}
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return headers, rows, err
		}
		rows = append(rows, rec)
	}
	return headers, rows, nil
}

func allowedFieldNames(modelInst orm.Model) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, f := range modelInst.Fields() {
		if f.Name == "" || f.Name == "id" {
			continue
		}
		allowed[f.Name] = struct{}{}
	}
	allowed["id"] = struct{}{}
	return allowed
}

func coerceCSVValue(raw string) interface{} {
	value := strings.TrimSpace(raw)
	if value == "" {
		return value
	}
	switch strings.ToLower(value) {
	case "true":
		return true
	case "false":
		return false
	}
	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		return n
	}
	if n, err := strconv.ParseFloat(value, 64); err == nil {
		return n
	}
	return value
}

func rowValuesFromMapping(headers, record []string, mapping map[string]string) map[string]interface{} {
	values := map[string]interface{}{}
	for i, col := range headers {
		field := strings.TrimSpace(mapping[col])
		if field == "" || field == "-" {
			continue
		}
		if i >= len(record) {
			continue
		}
		values[field] = coerceCSVValue(record[i])
	}
	return values
}
