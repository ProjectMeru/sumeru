package web

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

const contentRoutePrefix = "/web/content/"

// ContentHandler GET /web/content/{id} or /web/content/{model}/{field}/{resId}
func ContentHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	spec, ok := parseContentPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	uid := AuthenticatedUserID(r)
	ctx := r.Context()

	var rec map[string]interface{}
	var err error
	if spec.attachmentID > 0 {
		rec, err = orm.ResolveContentAttachment(ctx, spec.attachmentID, "", "", 0)
	} else {
		rec, err = orm.ResolveContentAttachment(ctx, 0, spec.resModel, spec.resField, spec.resID)
	}
	if err != nil || 0 == len(rec) {
		http.NotFound(w, r)
		return
	}
	if err := orm.CanReadAttachmentContent(ctx, int(uid), rec); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var reader io.ReadCloser
	storeFname := strings.TrimSpace(orm.AsString(rec["store_fname"]))
	if storeFname != "" {
		reader, err = orm.OpenAttachment(ctx, storeFname)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer reader.Close()
	} else {
		raw := orm.AsString(rec["datas"])
		data, decErr := base64.StdEncoding.DecodeString(raw)
		if decErr != nil {
			data = []byte(raw)
		}
		if 0 == len(data) {
			http.NotFound(w, r)
			return
		}
		reader = io.NopCloser(bytes.NewReader(data))
		defer reader.Close()
	}

	mimeType := strings.TrimSpace(orm.AsString(rec["mimetype"]))
	name := strings.TrimSpace(orm.AsString(rec["name"]))
	if name == "" {
		name = "download"
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	inline := strings.TrimSpace(r.URL.Query().Get("download")) == "0" ||
		strings.EqualFold(r.URL.Query().Get("download"), "false")
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if inline && contentInlineAllowed(mimeType) {
		w.Header().Set("Content-Disposition", safeContentDispositionInline(name))
	} else {
		w.Header().Set("Content-Disposition", safeContentDispositionFilename(name))
	}
	_, _ = io.Copy(w, reader)
}

type contentPathSpec struct {
	attachmentID int
	resModel     string
	resField     string
	resID        int
}

func parseContentPath(path string) (contentPathSpec, bool) {
	path = strings.TrimPrefix(path, contentRoutePrefix)
	path = strings.Trim(path, "/")
	if path == "" {
		return contentPathSpec{}, false
	}
	parts := strings.Split(path, "/")
	switch len(parts) {
	case 1:
		id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil || 0 >= id {
			return contentPathSpec{}, false
		}
		return contentPathSpec{attachmentID: id}, true
	case 2:
		// Legacy: field + parent record id requires ?model=
		return contentPathSpec{}, false
	case 3:
		resModel := strings.TrimSpace(parts[0])
		resField := strings.TrimSpace(parts[1])
		resID, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if resModel == "" || resField == "" || err != nil || 0 >= resID {
			return contentPathSpec{}, false
		}
		return contentPathSpec{resModel: resModel, resField: resField, resID: resID}, true
	default:
		return contentPathSpec{}, false
	}
}

func contentInlineAllowed(mimeType string) bool {
	m := strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	if m == "image/svg+xml" {
		return false
	}
	if strings.HasPrefix(m, "image/") {
		return true
	}
	return m == "application/pdf"
}
