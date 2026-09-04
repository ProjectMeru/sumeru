package report

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"html"
	"strings"
)

// ExportXLSX builds a minimal Office Open XML spreadsheet (.xlsx) for list export.
func ExportXLSX(ctx context.Context, in ExportCSVInput) ([]byte, error) {
	fields, err := ValidateFields(in.Model, in.Fields)
	if err != nil {
		return nil, err
	}
	rows, err := FetchRows(ctx, in.Model, in.Domain, in.RecordID)
	if err != nil {
		return nil, err
	}
	labels := FieldLabels(in.Model, fields)
	header := make([]string, len(fields))
	for i, f := range fields {
		if l := labels[f]; l != "" {
			header[i] = l
		} else {
			header[i] = f
		}
	}
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		line := make([]string, len(fields))
		for i, f := range fields {
			line[i] = formatCell(ctx, in.Model, f, row[f])
		}
		data = append(data, line)
	}
	return writeMinimalXLSX(header, data)
}

// ExportXLSXForTest builds a workbook without ORM access (tests only).
func ExportXLSXForTest(header []string, rows [][]string) ([]byte, error) {
	return writeMinimalXLSX(header, rows)
}

// WriteCSVForTest builds CSV bytes without ORM access (tests only).
func WriteCSVForTest(headers []string, rows [][]string) ([]byte, error) {
	return writeCSV(headers, rows)
}

// SheetXMLForTest builds worksheet XML without zipping (tests only).
func SheetXMLForTest(header []string, rows [][]string) string {
	return sheetXML(header, rows)
}

func writeMinimalXLSX(header []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"xl/workbook.xml":     workbookXML,
		"xl/_rels/workbook.xml.rels": workbookRelsXML,
		"xl/worksheets/sheet1.xml":   sheetXML(header, rows),
	}
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sheetXML(header []string, rows [][]string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	writeSheetRow(&b, 1, header)
	for i, row := range rows {
		writeSheetRow(&b, i+2, row)
	}
	b.WriteString(`</sheetData></worksheet>`)
	return b.String()
}

func writeSheetRow(b *strings.Builder, rowNum int, cells []string) {
	fmt.Fprintf(b, `<row r="%d">`, rowNum)
	for col, val := range cells {
		cellRef := cellRef(col, rowNum)
		esc := escapeOOXMLCell(sanitizeSpreadsheetCell(val))
		fmt.Fprintf(b, `<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, cellRef, esc)
	}
	b.WriteString(`</row>`)
}

func escapeOOXMLCell(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r == 0x9 || r == 0xA || r == 0xD || r >= 0x20 {
			b.WriteRune(r)
		}
	}
	return html.EscapeString(b.String())
}

func cellRef(col, row int) string {
	return columnName(col) + fmt.Sprint(row)
}

func columnName(col int) string {
	col++
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+col%26)) + name
		col /= 26
	}
	return name
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const workbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="Export" sheetId="1" r:id="rId1"/></sheets>
</workbook>`

const workbookRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`
