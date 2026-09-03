package api_test

import (
	"sumeru/core/server/api"
	"errors"
	"testing"
)



func TestClassify(t *testing.T) {
	cases := []struct {
		msg  string
		code string
	}{
		{"invalid json: boom", api.CodeInvalidJSON},
		{"invalid params: x", api.CodeInvalidJSON},
		{"empty body", api.CodeInvalidJSON},
		{"args: bad", api.CodeInvalidArgs},
		{"write requires args[0] ids", api.CodeInvalidArgs},
		{"kwargs must be object", api.CodeInvalidArgs},
		{"model is required", api.CodeValidationError},
		{"method is required", api.CodeValidationError},
		{"unknown model x", api.CodeModelNotFound},
		{"model foo not registered", api.CodeModelNotFound},
		{"not a public rpc method", api.CodeMethodNotAllowed},
		{"unsupported method x", api.CodeMethodNotAllowed},
		{"authentication required", api.CodeUnauthorized},
		{"access denied", api.CodeAccessDenied},
		{"record(s) not found", api.CodeNotFound},
		{"widget not found", api.CodeNotFound},
		{"something else", api.CodeInternalError},
	}
	for _, tc := range cases {
		code, _ := api.Classify(errors.New(tc.msg))
		if code != tc.code {
			t.Fatalf("api.Classify(%q) = %q; want %q", tc.msg, code, tc.code)
		}
	}
	if code, _ := api.Classify(nil); code != api.CodeInternalError {
		t.Fatalf("api.Classify(nil) = %q", code)
	}
}

func TestCodedError(t *testing.T) {
	err := api.NewRPCErrorForTest(api.CodeInvalidArgs, "bad args", map[string]interface{}{"n": 1})
	if err.Error() != "bad args" {
		t.Fatalf("Error() = %q", err.Error())
	}
}
