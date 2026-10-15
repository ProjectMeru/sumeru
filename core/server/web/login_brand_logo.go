package web

import (
	"encoding/base64"
	"net/http"
	"strings"
)

const maxLoginLogoDecodedBytes = 512 << 10 // 512 KiB

// LoginBrandLogoHandler serves the first company's login_logo for unauthenticated sign-in pages.
func LoginBrandLogoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload := loadLoginCompanyLogoPayload(r.Context())
	if payload == "" {
		http.NotFound(w, r)
		return
	}
	mime, data, ok := parseLoginLogoDataURL(payload)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if len(data) > maxLoginLogoDecodedBytes {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

func parseLoginLogoDataURL(raw string) (mime string, data []byte, ok bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "data:") {
		return "", nil, false
	}
	rest := strings.TrimPrefix(raw, "data:")
	semi := strings.Index(rest, ";")
	if 0 >= semi {
		return "", nil, false
	}
	mime = strings.TrimSpace(rest[:semi])
	if !strings.HasPrefix(mime, "image/") {
		return "", nil, false
	}
	if !strings.HasPrefix(rest[semi:], ";base64,") {
		return "", nil, false
	}
	b64 := strings.TrimPrefix(rest[semi:], ";base64,")
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", nil, false
	}
	if 0 == len(decoded) {
		return "", nil, false
	}
	return mime, decoded, true
}

func registerLoginBrandRoutes() {
	registerPublic(http.MethodGet, loginBrandLogoRoute, LoginBrandLogoHandler)
}
