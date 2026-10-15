package orm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// AttachmentScanner runs before bytes are written to the filestore (optional AV integration).
type AttachmentScanner func(ctx context.Context, name, mime string, body io.Reader) error

var attachmentScanner AttachmentScanner

// SetAttachmentScanner registers a virus-scan hook; nil clears it.
func SetAttachmentScanner(fn AttachmentScanner) {
	attachmentScanner = fn
}

// CreateBinaryAttachmentInput stores blob in filestore and creates sys.attachment.
type CreateBinaryAttachmentInput struct {
	Name      string
	ResModel  string
	ResField  string
	ResID     int
	Data      []byte
	Mimetype  string
	CompanyID int64
}

// CreateBinaryAttachment validates MIME, optional scan, filestore write, ORM create.
func CreateBinaryAttachment(ctx context.Context, in CreateBinaryAttachmentInput) (int, error) {
	if 0 == len(in.Data) {
		return 0, fmt.Errorf("attachment: empty data")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "file"
	}
	mime := normalizeAttachmentMIME(strings.TrimSpace(in.Mimetype), in.Data)
	if err := ValidateAttachmentMIME(mime, in.Data); err != nil {
		return 0, err
	}
	if attachmentScanner != nil {
		if err := attachmentScanner(ctx, name, mime, bytes.NewReader(in.Data)); err != nil {
			return 0, fmt.Errorf("attachment scan: %w", err)
		}
	}
	storeKey := fmt.Sprintf("%s_%d_%s", strings.TrimSpace(in.ResModel), in.ResID, time.Now().UTC().Format("20060102T150405.000"))
	storeFname, size, err := StoreAttachment(ctx, storeKey, in.Data)
	if err != nil {
		return 0, err
	}
	vals := map[string]interface{}{
		"name":        name,
		"mimetype":    mime,
		"file_size":   size,
		"store_fname": storeFname,
	}
	if m := strings.TrimSpace(in.ResModel); m != "" {
		vals["model"] = m
	}
	if in.ResID > 0 {
		vals["res_id"] = in.ResID
	}
	if f := strings.TrimSpace(in.ResField); f != "" {
		vals["res_field"] = f
	}
	if in.CompanyID > 0 {
		vals["company_id"] = in.CompanyID
	} else if cid := CompanyIDFromContext(ctx); cid > 0 {
		vals["company_id"] = cid
	}
	inst, ok := Registry["sys.attachment"]
	if !ok {
		return 0, fmt.Errorf("sys.attachment not registered")
	}
	return Create(ctx, inst, vals)
}
