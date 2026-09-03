package server

import "os"

// StripLeadingArgsSeparator removes a leading "--" from os.Args (Makefile / go run convention).
func StripLeadingArgsSeparator() {
	if len(os.Args) > 1 && os.Args[1] == "--" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}
}
