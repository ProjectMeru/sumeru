package web_test

import (
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestParsePinnedAppsJSONBody(t *testing.T) {
	body := `{"modules":["mail","contacts"]}`
	req := httptest.NewRequest("POST", web.TestPinnedAppsRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	request, ok := web.ParsePinnedAppsJSONBody(req)
	if !ok || len(request.Modules) != 2 || request.Modules[0] != "mail" {
		t.Fatalf("unexpected request: %+v ok=%v", request, ok)
	}
}

func TestParsePinnedAppsFormBody(t *testing.T) {
	req := httptest.NewRequest("POST", web.TestPinnedAppsRoute, strings.NewReader("modules=%5B%22sale%22%5D"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	request, ok := web.ParsePinnedAppsFormBody(req)
	if !ok || len(request.Modules) != 1 || request.Modules[0] != "sale" {
		t.Fatalf("unexpected request: %+v ok=%v", request, ok)
	}
}

func TestDecodePinnedModulesFieldEmpty(t *testing.T) {
	modules, ok := web.DecodePinnedModulesField("")
	if !ok || len(modules) != 0 {
		t.Fatalf("got %v ok=%v", modules, ok)
	}
}

func TestDecodePinnedModulesFieldInvalidJSON(t *testing.T) {
	_, ok := web.DecodePinnedModulesField("{bad")
	if ok {
		t.Fatal("invalid json should fail")
	}
}

func TestNormalizePinnedModules(t *testing.T) {
	if got := web.NormalizePinnedModules(nil); len(got) != 0 {
		t.Fatalf("nil should become empty slice, got %v", got)
	}
}

func TestParsePinnedAppsRequestUsesJSONContentType(t *testing.T) {
	req := httptest.NewRequest("POST", web.TestPinnedAppsRoute, strings.NewReader(`{"modules":["crm"]}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	request, ok := web.ParsePinnedAppsRequest(req)
	if !ok || len(request.Modules) != 1 || request.Modules[0] != "crm" {
		t.Fatalf("unexpected request: %+v ok=%v", request, ok)
	}
}
