package web_test

import (
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestParseCompanySwitchForm(t *testing.T) {
	body := "company_id=42&next=%2Fweb%2Fhome"
	req := httptest.NewRequest("POST", web.TestCompanySwitchRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	form := web.ParseCompanySwitchForm(req)
	if form.CompanyID != 42 || form.Next != "/web/home" {
		t.Fatalf("unexpected form: %+v", form)
	}
}

func TestParseCompanySwitchFormInvalidCompanyID(t *testing.T) {
	req := httptest.NewRequest("POST", web.TestCompanySwitchRoute, strings.NewReader("company_id=abc"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	form := web.ParseCompanySwitchForm(req)
	if form.CompanyID != 0 {
		t.Fatalf("expected company id 0, got %d", form.CompanyID)
	}
}
