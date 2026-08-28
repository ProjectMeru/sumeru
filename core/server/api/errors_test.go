package api

import (
	"errors"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		msg  string
		code string
	}{
		{"invalid json: boom", CodeInvalidJSON},
		{"invalid params: x", CodeInvalidJSON},
		{"empty body", CodeInvalidJSON},
		{"args: bad", CodeInvalidArgs},
		{"write requires args[0] ids", CodeInvalidArgs},
		{"kwargs must be object", CodeInvalidArgs},
		{"model is required", CodeValidationError},
		{"method is required", CodeValidationError},
		{"unknown model x", CodeModelNotFound},
		{"model foo not registered", CodeModelNotFound},
		{"not a public rpc method", CodeMethodNotAllowed},
		{"unsupported method x", CodeMethodNotAllowed},
		{"authentication required", CodeUnauthorized},
		{"access denied", CodeAccessDenied},
		{"record(s) not found", CodeNotFound},
		{"widget not found", CodeNotFound},
		{"something else", CodeInternalError},
	}
	for _, tc := range cases {
		code, _ := Classify(errors.New(tc.msg))
		if code != tc.code {
			t.Fatalf("Classify(%q) = %q; want %q", tc.msg, code, tc.code)
		}
	}
	if code, _ := Classify(nil); code != CodeInternalError {
		t.Fatalf("Classify(nil) = %q", code)
	}
}

func TestCodedError(t *testing.T) {
	err := newRPCError(CodeInvalidArgs, "bad args", map[string]interface{}{"n": 1})
	if err.Error() != "bad args" {
		t.Fatalf("Error() = %q", err.Error())
	}
}
