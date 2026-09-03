// Package acceptance holds platform parity acceptance checks per workstream.
// Run: go test ./test/acceptance/...
package acceptance

import (
	"testing"
)

// Workstream smoke tests — each area has 3–5 checks referenced by the Enterprise parity plan.

func TestWorkstreamA_SWCClient(t *testing.T) {
	if !swcViewTypesRegistered() {
		t.Fatal("expected list, form, kanban, graph, pivot view types in SWC registry")
	}
	if !sumTemplateCompiles() {
		t.Fatal("sum-template compiler should accept t-if/t-foreach")
	}
}

func TestWorkstreamB_Reporting(t *testing.T) {
	if !reportCapabilitiesParse() {
		t.Fatal("report capabilities should parse csv,pdf,xlsx download attrs")
	}
	if !templatePDFGenerates() {
		t.Fatal("template PDF export should produce bytes")
	}
}

func TestWorkstreamC_Security(t *testing.T) {
	if !recordRuleDomainCompiles() {
		t.Fatal("record rule domain should compile to SQL fragment")
	}
	if !fieldAccessModelRegistered() {
		t.Fatal("sys.field.access model should be registered")
	}
}

func TestWorkstreamD_Views(t *testing.T) {
	if !graphViewUsesReadGroup() {
		t.Fatal("graph view should call read_group aggregation")
	}
	if !mapViewHasEmbeddedMap() {
		t.Fatal("map view should render embedded map canvas")
	}
}

func TestWorkstreamE_XMLEngine(t *testing.T) {
	if !xpathInheritApplies() {
		t.Fatal("xpath inherit should apply field insertion")
	}
	if !safeEvalParsesDomain() {
		t.Fatal("SafeEval should parse domain list literals")
	}
	if !xpathHasClassSupported() {
		t.Fatal("xpath v2 should support hasclass() predicates")
	}
}

func TestWorkstreamF_Filters(t *testing.T) {
	if !domainOperatorsSupported() {
		t.Fatal("domain should support in, not in, ilike")
	}
	if !savedSearchModelExists() {
		t.Fatal("swc.saved.search model should exist")
	}
}

func TestWorkstreamG_ORM(t *testing.T) {
	if !readGroupAggregates() {
		t.Fatal("read_group should support sum and count")
	}
	if !modelInheritTagSupported() {
		t.Fatal("model inherit= tag should produce extend spec")
	}
}

func TestWorkstreamH_Relations(t *testing.T) {
	if !x2mCommandsSupported() {
		t.Fatal("x2m ORM should accept standard command tuples")
	}
}

func TestWorkstreamI_Debug(t *testing.T) {
	if !devFeaturesParse() {
		t.Fatal("dev_features INI should accept sql,access flags")
	}
}

func TestWorkstreamJ_Extensibility(t *testing.T) {
	if !attachmentFilestoreAvailable() {
		t.Fatal("sys.attachment filestore should store and retrieve blobs")
	}
}
