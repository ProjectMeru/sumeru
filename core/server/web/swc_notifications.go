package web

import (
	"net/http"
)

const (
	swcNotificationsRoute     = "/web/swc/notifications"
	swcNotificationsReadRoute = "/web/swc/notifications/read"
)

func registerSwcNotificationRoutes() {
	registerSession(http.MethodGet, swcNotificationsRoute, SwcNotificationsHandler)
	registerSession(http.MethodPost, swcNotificationsReadRoute, SwcNotificationsMarkReadHandler)
}

type swcNotificationItem struct {
	ID          int    `json:"id"`
	MessageID   int    `json:"messageId"`
	Body        string `json:"body"`
	Author      string `json:"author"`
	RecordModel string `json:"recordModel"`
	RecordID    int    `json:"recordId"`
	IsRead      bool   `json:"isRead"`
	CreateDate  string `json:"createDate"`
}

// SwcNotificationsHandler GET /web/swc/notifications
func SwcNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	writeJSONResponse(w, map[string]interface{}{
		"items":  []swcNotificationItem{},
		"unread": 0,
	})
}

// SwcNotificationsMarkReadHandler POST /web/swc/notifications/read
func SwcNotificationsMarkReadHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !ParsePostForm(w, r) {
		return
	}
	writeJSONResponse(w, map[string]interface{}{"ok": true})
}
