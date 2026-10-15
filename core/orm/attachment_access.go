package orm

import (
	"context"
	"fmt"
	"strings"
)

// CanReadAttachmentContentForTest exposes attachment ACL for external tests.
func CanReadAttachmentContentForTest(ctx context.Context, uid int, att map[string]interface{}) error {
	return CanReadAttachmentContent(ctx, uid, att)
}

// CanReadAttachmentContent checks sys.attachment read plus linked record read when set.
func CanReadAttachmentContent(ctx context.Context, uid int, att map[string]interface{}) error {
	if err := CheckModelAccess(ctx, uid, "sys.attachment", "read"); err != nil {
		return err
	}
	resModel := strings.TrimSpace(AsString(att["model"]))
	resID, _ := CoerceInt64(att["res_id"])
	if resModel == "" || 0 >= resID {
		return nil
	}
	if _, ok := Registry[resModel]; !ok {
		return fmt.Errorf("unknown linked model %q", resModel)
	}
	if err := CheckModelAccess(ctx, uid, resModel, "read"); err != nil {
		return err
	}
	row, err := SearchOne(ctx, resModel, map[string]interface{}{"id": resID})
	if err != nil {
		return fmt.Errorf("linked record not found")
	}
	if 0 == len(row) {
		return fmt.Errorf("linked record not found")
	}

	return nil
}

// ResolveContentAttachment loads sys.attachment for content serving.
// attachmentID > 0 loads by primary key; otherwise model+resField+resID must identify a row.
// ResolveContentAttachmentForTest exposes attachment lookup for external tests.
func ResolveContentAttachmentForTest(ctx context.Context, attachmentID int, resModel, resField string, resID int) (map[string]interface{}, error) {
	return ResolveContentAttachment(ctx, attachmentID, resModel, resField, resID)
}

func ResolveContentAttachment(ctx context.Context, attachmentID int, resModel, resField string, resID int) (map[string]interface{}, error) {
	if attachmentID > 0 {
		return SearchOne(ctx, "sys.attachment", map[string]interface{}{"id": attachmentID})
	}
	resModel = strings.TrimSpace(resModel)
	resField = strings.TrimSpace(resField)
	if resModel == "" || resField == "" || 0 >= resID {
		return nil, fmt.Errorf("attachment not found")
	}
	domain := map[string]interface{}{
		"model":     resModel,
		"res_id":    resID,
		"res_field": resField,
	}
	return SearchOne(ctx, "sys.attachment", domain)
}
