package applog_test

import (
	"testing"
	"time"

	"sumeru/core/applog"
)

func TestParseLogTimezone_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
	}{
		{"UTC"},
		{"Local"},
		{"America/New_York"},
		{"Invalid/Zone/XYZ"},
		{""},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			loc, name := applog.ParseLogTimezone(tc.input)
			if loc == nil {
				t.Fatalf("ParseLogTimezone(%q) nil location", tc.input)
			}
			if tc.input == "Invalid/Zone/XYZ" && loc != time.UTC {
				t.Fatalf("invalid zone should fall back to UTC, got %v", loc)
			}
			applog.SetLogLocationForTest(loc)
			if got := applog.EffectiveLocationForTest(); got != loc {
				t.Fatalf("EffectiveLocationForTest mismatch")
			}
			_ = name
		})
	}
}

func TestEffectiveLocation_default(t *testing.T) {
	applog.SetLogLocationForTest(nil)
	loc := applog.EffectiveLocationForTest()
	if loc == nil {
		t.Fatal("expected non-nil default location")
	}
	if loc.String() == "" {
		t.Fatal("empty location name")
	}
	_ = time.Now().In(loc)
}
