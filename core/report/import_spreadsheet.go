package report

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ParseUploadContentForTest exposes upload normalization for external tests.
func ParseUploadContentForTest(filename string, content []byte) ([]byte, error) {
	return ParseUploadContent(filename, content)
}

// ParseUploadContent normalizes CSV or first-sheet XLSX into CSV bytes.
func ParseUploadContent(filename string, content []byte) ([]byte, error) {
	name := strings.ToLower(strings.TrimSpace(filename))
	if strings.HasSuffix(name, ".xlsx") {
		return xlsxFirstSheetCSV(content)
	}
	return content, nil
}

func xlsxFirstSheetCSV(content []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("xlsx: %w", err)
	}
	var shared []string
	var sheetXML []byte
	for _, f := range zr.File {
		switch f.Name {
		case "xl/sharedStrings.xml":
			data, err := readZipEntry(f)
			if err != nil {
				return nil, err
			}
			shared = parseSharedStrings(data)
		case "xl/worksheets/sheet1.xml":
			sheetXML, err = readZipEntry(f)
			if err != nil {
				return nil, err
			}
		}
	}
	if 0 == len(sheetXML) {
		return nil, fmt.Errorf("xlsx: sheet1 missing")
	}
	rows := parseSheetRows(sheetXML, shared)
	if 0 == len(rows) {
		return nil, fmt.Errorf("xlsx: empty sheet")
	}
	return writeCSV(rows[0], rows[1:])
}

func readZipEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func parseSharedStrings(data []byte) []string {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var out []string
	var cur strings.Builder
	inSI := false
	inT := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out
		}
		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "si" {
				inSI = true
				cur.Reset()
			}
			if inSI && el.Name.Local == "t" {
				inT = true
			}
		case xml.CharData:
			if inT {
				cur.Write(el)
			}
		case xml.EndElement:
			if el.Name.Local == "t" {
				inT = false
			}
			if el.Name.Local == "si" {
				out = append(out, cur.String())
				inSI = false
			}
		}
	}
	return out
}

func parseSheetRows(data []byte, shared []string) [][]string {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var rows [][]string
	var row []string
	inRow := false
	inValue := false
	cellType := ""
	var val strings.Builder
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rows
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "row":
				inRow = true
				row = nil
			case "c":
				cellType = ""
				for _, a := range el.Attr {
					if a.Name.Local == "t" {
						cellType = a.Value
					}
				}
				val.Reset()
			case "v":
				inValue = true
			}
		case xml.CharData:
			if inValue {
				val.Write(el)
			}
		case xml.EndElement:
			switch el.Name.Local {
			case "v":
				inValue = false
				text := val.String()
				if cellType == "s" {
					if idx, err := parseInt(text); err == nil && idx >= 0 && idx < len(shared) {
						text = shared[idx]
					}
				}
				row = append(row, text)
			case "row":
				if inRow && len(row) > 0 {
					rows = append(rows, row)
				}
				inRow = false
			}
		}
	}
	return rows
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
