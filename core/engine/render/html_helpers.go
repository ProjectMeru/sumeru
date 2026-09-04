package render

import (
	"html/template"
	"strings"
	"unicode"

	"sumeru/core/engine/parser"
)

// SafeImageSrc reports whether src is safe to embed in an img tag (http(s), data URL, or site-relative).
func SafeImageSrc(src string) bool {
	src = strings.TrimSpace(src)
	return src != "" && (strings.HasPrefix(src, "http://") ||
		strings.HasPrefix(src, "https://") ||
		strings.HasPrefix(src, "data:") ||
		strings.HasPrefix(src, "/"))
}

// SafeIframeURL reports whether src is safe for an iframe (https or site-relative path).
// Rejects javascript:, data:, file:, and protocol-relative URLs.
func SafeIframeURL(src string) bool {
	src = strings.TrimSpace(src)
	if src == "" {
		return false
	}
	lower := strings.ToLower(src)
	if strings.HasPrefix(lower, "javascript:") ||
		strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "file:") ||
		strings.HasPrefix(lower, "vbscript:") ||
		strings.HasPrefix(src, "//") {
		return false
	}
	if strings.HasPrefix(src, "/") {
		return true
	}
	return strings.HasPrefix(lower, "https://")
}

// SafeIframeURLAllowHTTP is SafeIframeURL plus http:// absolute URLs (dev-only callers).
func SafeIframeURLAllowHTTP(src string) bool {
	if SafeIframeURL(src) {
		return true
	}
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(src)), "http://")
}

// FieldDisplayLabel returns the column/field label from XML string attr or a humanized field name.
func FieldDisplayLabel(field parser.Field) string {
	if label := strings.TrimSpace(field.Label); label != "" {
		return label
	}
	words := strings.Fields(strings.ReplaceAll(field.Name, "_", " "))
	for i, word := range words {
		if word == "" {
			continue
		}
		runes := []rune(strings.ToLower(word))
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// writeSaveCancelButtons renders form Save/Cancel controls for workspace record toolbar.
func writeSaveCancelButtons(sb *strings.Builder, cancelURL string) {
	sb.WriteString(`<button type="submit" form="sum-workspace-record-form" class="sum-list-btn-new">Save</button>`)
	sb.WriteString(`<a href="` + template.HTMLEscapeString(cancelURL) + `" class="sum-list-btn-ghost">Cancel</a>`)
}
