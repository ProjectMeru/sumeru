package importgen_test

import (
	"testing"

	"sumeru/core/modelmeta"
)

func TestModelNameFromGo(t *testing.T) {
	tests := map[string]string{
		"CoreCompany":       "core.company",
		"HrEmployee":        "hr.employee",
		"Partner":           "partner",
		"SysModuleCategory": "sys.module.category",
	}
	for in, want := range tests {
		if got := modelmeta.ModelNameFromGo(in); got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
}

func TestHeuristicGoName(t *testing.T) {
	tests := map[string]string{
		"core.partner":        "CorePartner",
		"hr.employee":         "HrEmployee",
		"core.company":        "CoreCompany",
		"sys.module.category": "SysModuleCategory",
	}
	for in, want := range tests {
		if got := modelmeta.HeuristicGoName(in); got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
}
