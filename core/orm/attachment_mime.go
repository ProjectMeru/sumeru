package orm

import (
	"fmt"
	"net/http"
	"strings"
)

var blockedAttachmentMIMEPrefixes = []string{
	"text/html",
	"application/javascript",
	"application/x-javascript",
	"text/javascript",
	"application/x-sh",
	"application/x-msdownload",
	"application/vnd.microsoft.portable-executable",
}

var allowedAttachmentMIMEPrefixes = []string{
	"image/",
	"text/plain",
	"text/csv",
	"application/pdf",
	"application/json",
	"application/xml",
	"text/xml",
	"application/zip",
	"application/gzip",
	"application/vnd.",
	"application/msword",
	"application/vnd.openxmlformats-officedocument",
	"application/vnd.oasis.opendocument",
	"application/octet-stream",
}

// ValidateAttachmentMIME rejects dangerous content types for uploads.
func ValidateAttachmentMIME(mimeType string, data []byte) error {
	mimeType = normalizeAttachmentMIME(mimeType, data)
	if mimeType == "" {
		return fmt.Errorf("attachment: unknown or disallowed content type")
	}
	lower := strings.ToLower(mimeType)
	for _, blocked := range blockedAttachmentMIMEPrefixes {
		if strings.HasPrefix(lower, blocked) {
			return fmt.Errorf("attachment: content type %q is not allowed", mimeType)
		}
	}
	for _, allowed := range allowedAttachmentMIMEPrefixes {
		if strings.HasPrefix(lower, allowed) {
			return nil
		}
	}
	return fmt.Errorf("attachment: content type %q is not allowed", mimeType)
}

func normalizeAttachmentMIME(declaredMIME string, data []byte) string {
	declaredMIME = strings.TrimSpace(strings.Split(declaredMIME, ";")[0])
	if declaredMIME != "" {
		return declaredMIME
	}
	if len(data) == 0 {
		return ""
	}
	sniffLen := len(data)
	if sniffLen > 512 {
		sniffLen = 512
	}
	return strings.TrimSpace(http.DetectContentType(data[:sniffLen]))
}
