package parser

import (
	"fmt"
	"strings"
)

func applyListOpenFlag(v *View) error {
	if v == nil || !strings.EqualFold(strings.TrimSpace(v.Type), "list") {
		return nil
	}
	open := strings.TrimSpace(v.ListOpenAttr)
	switch {
	case open == "", strings.EqualFold(open, "true"):
		v.ListNoRowOpen = false
	case strings.EqualFold(open, "false"):
		v.ListNoRowOpen = true
	default:
		err := fmt.Errorf("invalid open=%q (only \"true\" or \"false\" allowed)", open)
		if id, model := strings.TrimSpace(v.ID), strings.TrimSpace(v.Model); id != "" || model != "" {
			return fmt.Errorf("%w (view id=%q model=%q)", err, id, model)
		}
		return err
	}
	return nil
}
