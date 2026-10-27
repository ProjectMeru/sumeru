package im

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"sumeru/core/sdk"
)

const MessageModel = "im.message"

const (
	directUserDirectoryLimit = 80
	directUserSearchLimit    = 40
)

// SearchInternalUsers finds active internal users by name or login (excluding uid).
func SearchInternalUsers(ctx context.Context, uid int, query string, limit int) ([]map[string]interface{}, error) {
	if !sdk.DatabaseOpen() {
		return nil, nil
	}
	if err := sdk.CheckModelAccess(ctx, uid, "core.user", "read"); err != nil {
		return nil, err
	}
	q := strings.TrimSpace(query)
	if 0 >= limit {
		if q == "" {
			limit = directUserDirectoryLimit
		} else {
			limit = directUserSearchLimit
		}
	}
	if q == "" && limit > directUserDirectoryLimit {
		limit = directUserDirectoryLimit
	}
	if q != "" && limit > directUserSearchLimit {
		limit = directUserSearchLimit
	}
	domain := [][]interface{}{
		{"active", "=", true},
		{"user_type", "=", "internal"},
		{"id", "!=", uid},
	}
	for _, tok := range strings.Fields(q) {
		pattern := ilikePattern(tok)
		domain = append(domain, []interface{}{"|"})
		domain = append(domain, []interface{}{"name", "ilike", pattern})
		domain = append(domain, []interface{}{"login", "ilike", pattern})
	}
	return sdk.SearchPage(ctx, "core.user", domain, limit, 0, "name ASC")
}

func ilikePattern(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return "%"
	}
	if !strings.Contains(token, "%") {
		return "%" + token + "%"
	}
	return token
}

// ListDirectConversations returns recent peers for uid with last message preview.
func ListDirectConversations(ctx context.Context, uid int, limit int) ([]map[string]interface{}, error) {
	if !sdk.DatabaseOpen() {
		return nil, nil
	}
	if err := sdk.CheckModelAccess(ctx, uid, MessageModel, "read"); err != nil {
		return nil, err
	}
	if 0 >= limit || limit > 50 {
		limit = 30
	}
	domain := [][]interface{}{
		{"|"},
		{"sender_id", "=", uid},
		{"recipient_id", "=", uid},
	}
	rows, err := sdk.SearchPage(ctx, MessageModel, domain, 200, 0, "create_date DESC")
	if err != nil {
		return nil, err
	}
	type conv struct {
		preview string
		when    string
	}
	seen := map[int]conv{}
	order := []int{}
	for _, row := range rows {
		sid, _ := sdk.CoerceInt64(row["sender_id"])
		rid, _ := sdk.CoerceInt64(row["recipient_id"])
		var peer int
		switch uid {
		case int(sid):
			peer = int(rid)
		case int(rid):
			peer = int(sid)
		default:
			continue
		}
		if 0 >= peer || peer == uid {
			continue
		}
		if _, ok := seen[peer]; ok {
			continue
		}
		preview := strings.TrimSpace(sdk.AsString(row["body"]))
		if len(preview) > 80 {
			preview = preview[:77] + "…"
		}
		seen[peer] = conv{
			preview: preview,
			when:    strings.TrimSpace(sdk.AsString(row["create_date"])),
		}
		order = append(order, peer)
		if len(order) >= limit {
			break
		}
	}
	unreadByPeer := directUnreadBySender(ctx, uid)
	out := make([]map[string]interface{}, 0, len(order))
	for _, peerID := range order {
		c := seen[peerID]
		name := UserDisplayName(ctx, peerID)
		out = append(out, map[string]interface{}{
			"userId":      peerID,
			"name":        name,
			"preview":     c.preview,
			"lastMessage": c.when,
			"unreadCount": unreadByPeer[peerID],
		})
	}
	return out, nil
}

func directUnreadBySender(ctx context.Context, uid int) map[int]int {
	out := map[int]int{}
	rows, err := sdk.Search(ctx, MessageModel, [][]interface{}{
		{"recipient_id", "=", uid},
		{"is_read", "=", false},
	})
	if err != nil {
		return out
	}
	for _, row := range rows {
		sid, _ := sdk.CoerceInt64(row["sender_id"])
		if 0 >= sid {
			continue
		}
		out[int(sid)]++
	}
	return out
}

// ListDirectThread returns messages between uid and peerID oldest first.
func ListDirectThread(ctx context.Context, uid, peerID int) ([]map[string]interface{}, error) {
	if 0 >= peerID || peerID == uid {
		return nil, fmt.Errorf("invalid peer")
	}
	if !sdk.DatabaseOpen() {
		return nil, nil
	}
	if err := sdk.CheckModelAccess(ctx, uid, MessageModel, "read"); err != nil {
		return nil, err
	}
	if err := assertInternalUser(ctx, uid, peerID); err != nil {
		return nil, err
	}
	rowsA, err := sdk.SearchPage(ctx, MessageModel, [][]interface{}{
		{"sender_id", "=", uid},
		{"recipient_id", "=", peerID},
	}, 120, 0, "create_date ASC")
	if err != nil {
		return nil, err
	}
	rowsB, err := sdk.SearchPage(ctx, MessageModel, [][]interface{}{
		{"sender_id", "=", peerID},
		{"recipient_id", "=", uid},
	}, 120, 0, "create_date ASC")
	if err != nil {
		return nil, err
	}
	rows := append(rowsA, rowsB...)
	sortDirectMessages(rows)
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		msgID, _ := sdk.CoerceInt64(row["id"])
		sid, _ := sdk.CoerceInt64(row["sender_id"])
		resID, _ := sdk.CoerceInt64(row["res_id"])
		out = append(out, map[string]interface{}{
			"id":          int(msgID),
			"body":        strings.TrimSpace(sdk.AsString(row["body"])),
			"createDate":  strings.TrimSpace(sdk.AsString(row["create_date"])),
			"outgoing":    int(sid) == uid,
			"resModel":    strings.TrimSpace(sdk.AsString(row["res_model"])),
			"resId":       int(resID),
			"attachments": listMessageAttachments(ctx, uid, int(msgID)),
		})
	}
	_ = MarkDirectThreadRead(ctx, uid, peerID)
	return out, nil
}

func listMessageAttachments(ctx context.Context, uid, messageID int) []map[string]interface{} {
	if 0 >= messageID {
		return nil
	}
	if err := sdk.CheckModelAccess(ctx, uid, "sys.attachment", "read"); err != nil {
		return nil
	}
	rows, err := sdk.Search(ctx, "sys.attachment", [][]interface{}{
		{"model", "=", MessageModel},
		{"res_id", "=", messageID},
	})
	if err != nil {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		id, _ := sdk.CoerceInt64(row["id"])
		out = append(out, map[string]interface{}{
			"id":       int(id),
			"name":     strings.TrimSpace(sdk.AsString(row["name"])),
			"url":      fmt.Sprintf("/web/content/%d", id),
			"mimetype": strings.TrimSpace(sdk.AsString(row["mimetype"])),
		})
	}
	return out
}

// PostDirectMessage sends an internal direct message from uid to recipientID.
func PostDirectMessage(ctx context.Context, uid, recipientID int, body string) (int, error) {
	return PostDirectMessageLinked(ctx, uid, recipientID, body, "", 0)
}

// PostDirectMessageLinked optionally attaches a related record (res_model / res_id).
func PostDirectMessageLinked(ctx context.Context, uid, recipientID int, body, resModel string, resID int) (int, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return 0, fmt.Errorf("message required")
	}
	if err := sdk.CheckModelAccess(ctx, uid, MessageModel, "create"); err != nil {
		return 0, err
	}
	if err := assertInternalUser(ctx, uid, recipientID); err != nil {
		return 0, err
	}
	if !sdk.ModelRegistered(MessageModel) {
		return 0, fmt.Errorf("internal chat not available")
	}
	vals := map[string]interface{}{
		"sender_id":    uid,
		"recipient_id": recipientID,
		"body":         body,
		"create_date":  time.Now().UTC(),
		"is_read":      false,
	}
	resModel = strings.TrimSpace(resModel)
	if resModel != "" && resID > 0 {
		vals["res_model"] = resModel
		vals["res_id"] = resID
	}
	id, err := sdk.Create(ctx, MessageModel, vals)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func assertInternalUser(ctx context.Context, uid, peerID int) error {
	if err := sdk.CheckModelAccess(ctx, uid, "core.user", "read"); err != nil {
		return err
	}
	u, err := sdk.SearchOne(ctx, "core.user", map[string]interface{}{"id": peerID})
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if !sdk.AsBool(u["active"]) {
		return fmt.Errorf("user not available")
	}
	if strings.TrimSpace(sdk.AsString(u["user_type"])) != "internal" {
		return fmt.Errorf("direct messages are for internal users only")
	}
	return nil
}

func sortDirectMessages(rows []map[string]interface{}) {
	sort.Slice(rows, func(i, j int) bool {
		a := strings.TrimSpace(sdk.AsString(rows[i]["create_date"]))
		b := strings.TrimSpace(sdk.AsString(rows[j]["create_date"]))
		return a < b
	})
}

// UserDisplayName returns a label for a user id.
func UserDisplayName(ctx context.Context, userID int) string {
	u, err := sdk.SearchOne(ctx, "core.user", map[string]interface{}{"id": userID})
	if err != nil {
		return fmt.Sprintf("User #%d", userID)
	}
	if nm := strings.TrimSpace(sdk.AsString(u["name"])); nm != "" {
		return nm
	}
	if lg := strings.TrimSpace(sdk.AsString(u["login"])); lg != "" {
		return lg
	}
	return fmt.Sprintf("User #%d", userID)
}

// SortDirectMessagesForTest exposes sort for tests.
func SortDirectMessagesForTest(rows []map[string]interface{}) {
	sortDirectMessages(rows)
}
