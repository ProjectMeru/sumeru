package web_test

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestLoginURLWithReturn(t *testing.T) {
	returnTo := "/web/home?menu_id=1"
	got := web.LoginURLWithReturn(returnTo)
	want := web.TestLoginRoute + "?next=" + url.QueryEscape(returnTo)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{header: "Bearer secret-key", want: "secret-key"},
		{header: "bearer secret-key", want: "secret-key"},
		{header: "Basic abc", want: ""},
		{header: "", want: ""},
	}
	for _, test := range tests {
		if got := web.BearerToken(test.header); got != test.want {
			t.Fatalf("web.BearerToken(%q) = %q want %q", test.header, got, test.want)
		}
	}
}

func TestParseLoginCredentials(t *testing.T) {
	body := "login=admin%40example.com&password=secret&next=%2Fweb%2Fapps"
	req := httptest.NewRequest("POST", web.TestLoginRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	credentials := web.ParseLoginCredentials(req)
	if credentials.Login != "admin@example.com" || credentials.Password != "secret" || credentials.Next != "/web/apps" {
		t.Fatalf("unexpected credentials: %+v", credentials)
	}
}
