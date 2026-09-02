package eval

import "testing"

func TestSafeEvalDomainList(t *testing.T) {
	got, err := SafeEval(`[('active', '=', True), ('state', 'in', ['draft', 'done'])]`)
	if err != nil {
		t.Fatal(err)
	}
	domain, ok := got.([][]interface{})
	if !ok || len(domain) != 2 {
		t.Fatalf("domain parse: %v", got)
	}
}

func TestSafeEvalLiterals(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want interface{}
	}{
		{"True", true},
		{"false", false},
		{"42", int64(42)},
		{`"hello"`, "hello"},
		{"(4,5)", []interface{}{int64(4), int64(5)}},
	} {
		got, err := SafeEval(tc.in)
		if err != nil {
			t.Fatalf("SafeEval(%q): %v", tc.in, err)
		}
		switch want := tc.want.(type) {
		case []interface{}:
			gotSlice, ok := got.([]interface{})
			if !ok || len(gotSlice) != len(want) {
				t.Fatalf("SafeEval(%q) = %v want slice %v", tc.in, got, want)
			}
		default:
			if got != want {
				t.Fatalf("SafeEval(%q) = %v want %v", tc.in, got, want)
			}
		}
	}
}
