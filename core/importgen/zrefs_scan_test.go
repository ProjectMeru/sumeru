package importgen

import "testing"

func TestFilterRefsForUsage_prefersDefiningModelOverInherit(t *testing.T) {
	refs := []modelRef{
		{
			GoName:         "PartnerContacts",
			TechnicalModel: "core.partner",
			ImportPath:     "sumeru/addons/contacts/models",
			HeuristicName:  "CorePartner",
			PhantomName:    "CorePartner",
			IsExtend:       true,
		},
		{
			GoName:         "CorePartner",
			TechnicalModel: "core.partner",
			ImportPath:     "sumeru/addons/base/models",
			HeuristicName:  "CorePartner",
			UseAlias:       true,
			IsExtend:       false,
		},
	}
	used := map[string]struct{}{"CorePartner": {}}
	filtered := filterRefsForUsage(refs, used)
	if len(filtered) != 1 {
		t.Fatalf("got %d refs, want 1", len(filtered))
	}
	if filtered[0].GoName != "CorePartner" {
		t.Fatalf("got %q, want CorePartner defining model", filtered[0].GoName)
	}
}
