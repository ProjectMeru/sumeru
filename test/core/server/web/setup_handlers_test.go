package web_test

import (
	"net/http/httptest"
	"sumeru/core/server/web"
	"testing"

	"sumeru/core/orm"
)

func TestParseSetupInitRequest(t *testing.T) {
	body := []byte(`{"company_name":"Acme","lang":"en","admin_name":"Admin","email":"a@example.com","password":"secret","setup_token":"tok"}`)
	recorder := httptest.NewRecorder()
	payload, ok := web.ParseSetupInitRequest(recorder, body)
	if !ok {
		t.Fatal("expected valid setup payload")
	}
	if payload.CompanyName != "Acme" || payload.Email != "a@example.com" || payload.SetupToken != "tok" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestParseSetupInitRequestInvalidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	_, ok := web.ParseSetupInitRequest(recorder, []byte(`{`))
	if ok {
		t.Fatal("invalid json should fail")
	}
}

func TestToSetupAdminParams(t *testing.T) {
	params := web.ToSetupAdminParams(web.SetupInitRequest{
		CompanyName: "Acme",
		Lang:        "en_US",
		AdminName:   "Jane Admin",
		Email:       "jane@example.com",
		Password:    "pw",
	})
	want := orm.SetupAdminParams{
		CompanyName: "Acme",
		Lang:        "en_US",
		FullName:    "Jane Admin",
		Email:       "jane@example.com",
		Password:    "pw",
	}
	if params != want {
		t.Fatalf("got %+v want %+v", params, want)
	}
}

func TestBuildSetupPageData(t *testing.T) {
	pageData := web.BuildSetupPageData()
	if len(pageData.Stylesheets) == 0 {
		t.Fatalf("expected setup stylesheets, got %+v", pageData)
	}
}
