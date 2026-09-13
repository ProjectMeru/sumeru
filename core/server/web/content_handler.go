package web

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

const contentRoutePrefix = "/web/content/"

// ContentHandler GET /web/content/{field}/{id} — authenticated attachment download.
func ContentHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	_, recordID, ok := parseContentPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	uid := AuthenticatedUserID(r)
	ctx := r.Context()
	if err := orm.CheckModelAccess(ctx, uid, "sys.attachment", "read"); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rec, err := orm.SearchOne(ctx, "sys.attachment", map[string]interface{}{"id": recordID})
	if err != nil || len(rec) == 0 {
		http.NotFound(w, r)
		return
	}
	storeFname := orm.AsString(rec["store_fname"])
	if storeFname == "" {
		http.NotFound(w, r)
		return
	}
	reader, err := orm.OpenAttachment(ctx, storeFname)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer reader.Close()

	mimeType := strings.TrimSpace(orm.AsString(rec["mimetype"]))
	name := strings.TrimSpace(orm.AsString(rec["name"]))
	if name == "" {
		name = "download"
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", safeContentDispositionFilename(name))
	_, _ = io.Copy(w, reader)
}

func parseContentPath(path string) (field string, recordID int, ok bool) {
	path = strings.TrimPrefix(path, contentRoutePrefix)
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", 0, false
	}
	field = strings.TrimSpace(parts[0])
	id, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if field == "" || err != nil || id <= 0 {
		return "", 0, false
	}
	return field, id, true
}
