package orm

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// TranslationCSVHeader is the column order for sys.translation import/export.
var TranslationCSVHeader = []string{"lang", "src", "value", "module"}

// TranslationTableName returns the quoted table name for sys.translation.
func TranslationTableName() (string, error) {
	return ModelToTableName("sys.translation")
}

// ParseTranslationCSV reads a translation CSV and validates required columns.
func ParseTranslationCSV(r io.Reader) (header []string, rows [][]string, err error) {
	reader := csv.NewReader(r)
	all, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(all) < 2 {
		return all[0], nil, nil
	}
	header = all[0]
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, req := range TranslationCSVHeader {
		if _, ok := col[req]; !ok {
			return header, nil, fmt.Errorf("missing column %q", req)
		}
	}
	return header, all[1:], nil
}

// ImportTranslationsCSV inserts or updates rows from parsed CSV data.
func ImportTranslationsCSV(ctx context.Context, db *sql.DB, header []string, rows [][]string) (int, error) {
	tableName, err := TranslationTableName()
	if err != nil {
		return 0, err
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	imported := 0
	for _, row := range rows {
		if len(row) < len(header) {
			continue
		}
		lang := strings.TrimSpace(row[col["lang"]])
		src := strings.TrimSpace(row[col["src"]])
		val := row[col["value"]]
		mod := strings.TrimSpace(row[col["module"]])
		if lang == "" || src == "" {
			continue
		}
		_, err := db.ExecContext(ctx, `
			INSERT INTO "`+tableName+`" (lang, src, value, module)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (lang, src, module) DO UPDATE SET value = EXCLUDED.value`, lang, src, val, mod)
		if err != nil {
			return imported, err
		}
		imported++
	}
	return imported, nil
}

// ExportTranslationsCSV writes all sys.translation rows to outPath.
func ExportTranslationsCSV(ctx context.Context, db *sql.DB, outPath string) (int, error) {
	tableName, err := TranslationTableName()
	if err != nil {
		return 0, err
	}
	rows, err := db.QueryContext(ctx,
		`SELECT lang, src, value, module FROM "`+tableName+`" ORDER BY lang, module, src`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	f, err := os.Create(outPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(TranslationCSVHeader); err != nil {
		return 0, err
	}
	count := 0
	for rows.Next() {
		var lang, src, value, module string
		if err := rows.Scan(&lang, &src, &value, &module); err != nil {
			return count, err
		}
		if err := w.Write([]string{lang, src, value, module}); err != nil {
			return count, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return count, err
	}
	w.Flush()
	return count, w.Error()
}
