package base_test

import (
	"context"
	"strings"
	"testing"

	_ "sumeru/addons/base"
	"sumeru/core/orm"
)

func TestCoreUserObjectActionsRegistered(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	for _, method := range []string{"action_archive", "action_reset_password"} {
		_, err := orm.RunObjectAction(ctx, "core.user", 1, method, nil)
		if err != nil && strings.Contains(err.Error(), "unknown object action") {
			t.Fatalf("%s not registered: %v", method, err)
		}
	}
}
