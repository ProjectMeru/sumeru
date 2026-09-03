package eval_test

import (
	"strings"
	"testing"
	"time"

	"sumeru/core/engine/eval"
)

func TestSafeEval_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    interface{}
		wantErr bool
		check   func(t *testing.T, got interface{})
	}{
		{"empty", "", nil, false, nil},
		{"null", "null", nil, false, nil},
		{"ref tuple", "ref('base.main_partner')", []interface{}{"ref", "base.main_partner"}, false, nil},
		{"empty ref", "ref('')", nil, true, nil},
		{"time today", "time.today()", nil, false, func(t *testing.T, got interface{}) {
			d, ok := got.(time.Time)
			if !ok || d.Hour() != 0 {
				t.Fatalf("time.today(): %v", got)
			}
		}},
		{"time now", "time.now()", nil, false, func(t *testing.T, got interface{}) {
			if _, ok := got.(time.Time); !ok {
				t.Fatalf("time.now(): %T", got)
			}
		}},
		{"bad time", "time.yesterday()", nil, true, nil},
		{"domain nested list", `[('state','in',['draft','done'])]`, nil, false, func(t *testing.T, got interface{}) {
			d, ok := got.([][]interface{})
			if !ok || len(d) != 1 {
				t.Fatalf("domain: %v", got)
			}
		}},
		{"empty domain", "[]", nil, false, func(t *testing.T, got interface{}) {
			d, ok := got.([][]interface{})
			if !ok || len(d) != 0 {
				t.Fatalf("empty domain: %v", got)
			}
		}},
		{"float", "3.14", 3.14, false, nil},
		{"comma list", "a,b,c", []interface{}{"a", "b", "c"}, false, nil},
		{"raw fallback", "custom_token", "custom_token", false, nil},
		{"bad domain triple", "[(bad)]", nil, true, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := eval.SafeEval(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.check != nil {
				tc.check(t, got)
				return
			}
			if tc.want != nil {
				switch want := tc.want.(type) {
				case []interface{}:
					gotSlice, ok := got.([]interface{})
					if !ok || len(gotSlice) != len(want) {
						t.Fatalf("got %v want %v", got, want)
					}
					for i := range want {
						if gotSlice[i] != want[i] {
							t.Fatalf("[%d] got %v want %v", i, gotSlice[i], want[i])
						}
					}
				default:
					if got != want {
						t.Fatalf("got %v (%T) want %v", got, got, want)
					}
				}
			}
		})
	}
}

func TestMustSafeEval_table(t *testing.T) {
	t.Parallel()
	if _, err := eval.MustSafeEval("1+2"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("MustSafeEval arithmetic: %v", err)
	}
	v, err := eval.MustSafeEval("True")
	if err != nil || v != true {
		t.Fatalf("MustSafeEval literal: v=%v err=%v", v, err)
	}
	v, err = eval.MustSafeEval(`[('x','=',1)]`)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := v.([][]interface{}); !ok {
		t.Fatalf("MustSafeEval domain: %T", v)
	}
}
