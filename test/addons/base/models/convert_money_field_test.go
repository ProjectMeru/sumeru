package models_test

import (
	"context"
	"testing"
	"time"

	_ "sumeru/addons/base"
	"sumeru/core/modelreg"
	"sumeru/core/sdk"
)

// TestConvertMoneyFieldRejectsUntaggedField checks ConvertMoneyField against a
// real base model field that carries no currency=<field> tag.
func TestConvertMoneyFieldRejectsUntaggedField(t *testing.T) {
	_ = modelreg.ActivateAll([]string{"base"})
	_, err := sdk.ConvertMoneyField(context.Background(), "core.user", "name",
		map[string]interface{}{"name": "x"}, 1, time.Time{})
	if err == nil {
		t.Fatal("expected error: core.user.name is not currency-tagged")
	}
}
