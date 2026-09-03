package server_test

import (
	"testing"

	"sumeru/core/server"
)

func TestNormalizeManifestAssetRel(t *testing.T) {
	tests := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"static/css/app.css", "static/css/app.css", true},
		{"  static/js/widget.js  ", "static/js/widget.js", true},
		{"", "", false},
		{"../secrets.env", "", false},
		{"/static/css/app.css", "", false},
		{"C:/windows/path/app.js", "", false},
		{`\\server\share\app.js`, "", false},
	}
	for _, tc := range tests {
		got, ok := server.NormalizeManifestAssetRelForTest(tc.in)
		if ok != tc.wantOK || got != tc.want {
			t.Fatalf("normalize(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestManifestAssetPublicURL(t *testing.T) {
	got := server.ManifestAssetPublicURLForTest("my_addon", "static/js/widget.js")
	want := "/static/addon-asset/my_addon/static/js/widget.js"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
