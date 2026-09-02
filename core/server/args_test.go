package server

import (
	"os"
	"testing"
)

func TestStripLeadingArgsSeparator(t *testing.T) {
	orig := append([]string(nil), os.Args...)
	t.Cleanup(func() { os.Args = orig })

	os.Args = []string{"sumeru", "--", "-c", "sumeru.conf"}
	StripLeadingArgsSeparator()
	if len(os.Args) != 3 || os.Args[1] != "-c" {
		t.Fatalf("expected [-c sumeru.conf] after strip, got %v", os.Args)
	}

	os.Args = []string{"sumeru", "-c", "sumeru.conf"}
	StripLeadingArgsSeparator()
	if len(os.Args) != 3 || os.Args[1] != "-c" {
		t.Fatalf("unchanged args expected, got %v", os.Args)
	}
}
