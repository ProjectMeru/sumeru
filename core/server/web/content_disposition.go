package web

import (
	"mime"
	"strings"
)

// safeContentDispositionFilename returns an RFC 5987-safe attachment filename value.
func safeContentDispositionFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.NewReplacer("\r", "", "\n", "", `"`, "", `\`, "").Replace(name)
	if name == "" {
		name = "download"
	}
	return mime.FormatMediaType("attachment", map[string]string{"filename": name})
}
