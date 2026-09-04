package render

import (
	"strings"

	"sumeru/core/orm"
)

func splitFlashMessages(flashes []FlashMessage) (inline, toast []FlashMessage) {
	for _, f := range flashes {
		if f.ToastOnly {
			toast = append(toast, f)
			continue
		}
		inline = append(inline, f)
	}
	return inline, toast
}

func recStr(rec map[string]interface{}, name string) string {
	if rec == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(rec[name]))
}
