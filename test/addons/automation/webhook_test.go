package automation_test

import (
	"sumeru/addons/automation"
	"testing"
)

func TestValidateWebhookURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		url     string
		wantErr bool
	}{
		{"", true},
		{"javascript:alert(1)", true},
		{"ftp://example.com/x", true},
		{"http://127.0.0.1/hook", true},
		{"https://localhost/hook", true},
		{"http://169.254.169.254/latest/meta-data", true},
		{"https://192.168.1.1/hook", true},
		{"https://10.0.0.5/hook", true},
		{"https://8.8.8.8/hook", false},
	}
	for _, tc := range cases {
		err := automation.ValidateWebhookURLForTest(tc.url)
		if tc.wantErr && err == nil {
			t.Errorf("%q: want error", tc.url)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("%q: unexpected error: %v", tc.url, err)
		}
	}
}
