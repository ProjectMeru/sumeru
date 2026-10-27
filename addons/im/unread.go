package im

import (
	"context"

	"sumeru/core/sdk"
)

// UnreadDirectCount returns incoming messages not yet marked read for uid.
func UnreadDirectCount(ctx context.Context, uid int) (int, error) {
	if 0 >= uid || !sdk.DatabaseOpen() {
		return 0, nil
	}
	if err := sdk.CheckModelAccess(ctx, uid, MessageModel, "read"); err != nil {
		return 0, err
	}
	if !sdk.ModelRegistered(MessageModel) {
		return 0, nil
	}
	domain := [][]interface{}{
		{"recipient_id", "=", uid},
		{"is_read", "=", false},
	}
	rows, err := sdk.Search(ctx, MessageModel, domain)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// MarkDirectThreadRead sets is_read on incoming messages in a peer thread.
func MarkDirectThreadRead(ctx context.Context, uid, peerID int) error {
	if 0 >= uid || 0 >= peerID {
		return nil
	}
	if err := sdk.CheckModelAccess(ctx, uid, MessageModel, "write"); err != nil {
		return err
	}
	rows, err := sdk.Search(ctx, MessageModel, [][]interface{}{
		{"sender_id", "=", peerID},
		{"recipient_id", "=", uid},
		{"is_read", "=", false},
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		id, _ := sdk.CoerceInt64(row["id"])
		if 0 >= id {
			continue
		}
		if _, err := sdk.Write(ctx, MessageModel, []int{int(id)}, map[string]interface{}{"is_read": true}); err != nil {
			return err
		}
	}
	return nil
}
