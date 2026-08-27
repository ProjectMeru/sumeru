package modelmeta

import (
	"strings"
	"unicode"
)

// FieldNameFromGo converts a Go exported field name to a snake_case column name.
func FieldNameFromGo(name string) string {
	if name == "" {
		return ""
	}
	if name == "ID" {
		return "id"
	}
	if strings.HasSuffix(name, "ID") && len(name) > 2 {
		return FieldNameFromGo(name[:len(name)-2]) + "_id"
	}
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// LabelFromGo derives a human label from a Go field name.
func LabelFromGo(name string) string {
	snake := FieldNameFromGo(name)
	if snake == "" {
		return ""
	}
	parts := strings.Split(snake, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// ModelNameFromGo converts a struct type name to a dotted technical model name.
func ModelNameFromGo(typeName string) string {
	if typeName == "" {
		return ""
	}
	var segments []string
	start := 0
	for i := 1; i < len(typeName); i++ {
		if unicode.IsUpper(rune(typeName[i])) && i > start {
			segments = append(segments, strings.ToLower(typeName[start:i]))
			start = i
		}
	}
	segments = append(segments, strings.ToLower(typeName[start:]))
	return strings.Join(segments, ".")
}

// HeuristicGoName maps a dotted technical model name to a Go exported type name.
func HeuristicGoName(technicalModel string) string {
	technicalModel = strings.TrimSpace(technicalModel)
	if technicalModel == "" || technicalModel == "-" {
		return ""
	}
	parts := strings.Split(technicalModel, ".")
	var b strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		b.WriteString(titleWord(part))
	}
	return b.String()
}

func titleWord(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
