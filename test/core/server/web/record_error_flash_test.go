package web_test

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/orm"
	"sumeru/core/server/web"
)

func TestUserFacingRecordErrorRecordRule(t *testing.T) {
	title, body, details, fields := web.UserFacingRecordError("record_save", "crm.lead", &orm.RecordRuleError{Model: "crm.lead"})
	if title != "Save failed" || body == "" || details == "" {
		t.Fatalf("got title=%q body=%q details=%q", title, body, details)
	}
	if body != "This record could not be saved with your access rights." {
		t.Fatalf("unexpected body: %q", body)
	}
	if len(fields) != 0 {
		t.Fatalf("expected no field errors, got %v", fields)
	}
}

func TestUserFacingRecordErrorFieldValidation(t *testing.T) {
	err := &orm.FieldValidationError{
		Field:   "name",
		Label:   "Opportunity",
		Message: "Opportunity is required.",
	}
	title, body, details, fields := web.UserFacingRecordError("record_save", "crm.lead", err)
	if title != "Save failed" || body != "Opportunity is required." || details == "" {
		t.Fatalf("got title=%q body=%q details=%q", title, body, details)
	}
	if len(fields) != 1 || fields[0] != "name" {
		t.Fatalf("unexpected field errors: %v", fields)
	}
}

func TestUserFacingRecordErrorRequiredField(t *testing.T) {
	_, body, _, fields := web.UserFacingRecordError("record_save", "crm.lead", errors.New(`required field "name" missing on model crm.lead`))
	if body != "name is required." {
		t.Fatalf("unexpected body: %q", body)
	}
	if len(fields) != 1 || fields[0] != "name" {
		t.Fatalf("unexpected field errors: %v", fields)
	}
}

func TestRecordErrorFlashCookieRoundTrip(t *testing.T) {
	recorder := httptest.NewRecorder()
	web.SetRecordErrorFlashForTest(recorder, web.PageFlash{
		Kind:    "error",
		Title:   "Save failed",
		Body:    "Could not save.",
		Details: "record rule failed for model crm.lead",
	})
	if len(recorder.Result().Cookies()) == 0 {
		t.Fatal("expected cookie to be set")
	}

	req := httptest.NewRequest("GET", "/web", nil)
	for _, c := range recorder.Result().Cookies() {
		req.AddCookie(c)
	}
	consumeRecorder := httptest.NewRecorder()
	flash, ok := web.ConsumeRecordErrorFlashForTest(req, consumeRecorder)
	if !ok {
		t.Fatal("expected flash from cookie")
	}
	if flash.Kind != "error" || flash.Title != "Save failed" || flash.Details == "" {
		t.Fatalf("unexpected flash: %+v", flash)
	}

	req2 := httptest.NewRequest("GET", "/web", nil)
	for _, c := range consumeRecorder.Result().Cookies() {
		req2.AddCookie(c)
	}
	if _, ok := web.ConsumeRecordErrorFlashForTest(req2, httptest.NewRecorder()); ok {
		t.Fatal("expected cookie to be consumed once")
	}
}

func TestEnsureFormEditRedirectURLClearsIDOnCreate(t *testing.T) {
	got := web.EnsureFormEditRedirectURL("/web?action=1&view_type=form&id=0&edit=1", true)
	if strings.Contains(got, "id=") {
		t.Fatalf("expected id removed: %q", got)
	}
	if !strings.Contains(got, "edit=1") {
		t.Fatalf("expected edit=1: %q", got)
	}
}

func TestActionDefaultFieldValues(t *testing.T) {
	got := web.ActionDefaultFieldValues(map[string]interface{}{
		"context": `{"default_type":"opportunity","view_id":"42"}`,
	})
	if got["type"] != "opportunity" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
}
