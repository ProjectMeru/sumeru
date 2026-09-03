package module_test

import (
	"fmt"
	"testing"

	"sumeru/core/module"
)

func TestAllowMenuParseFallback(t *testing.T) {
	if !module.AllowMenuParseFallback(nil) {
		t.Fatal("nil error should allow menu fallback")
	}
	rootErr := fmt.Errorf("views/x.xml: module XML root must be <sumeru>, got <data>")
	if module.AllowMenuParseFallback(rootErr) {
		t.Fatal("invalid module root should not allow menu fallback")
	}
	unmarshalErr := fmt.Errorf("XML syntax error")
	if !module.AllowMenuParseFallback(unmarshalErr) {
		t.Fatal("unmarshal errors should allow menu fallback")
	}
}

func TestRecordsFromActionsEmpty(t *testing.T) {
	if module.RecordsFromActions(nil) != nil {
		t.Fatal("nil actions should return nil")
	}
	if len(module.RecordsFromActions(nil)) != 0 {
		t.Fatal("empty actions should return nil slice")
	}
}
