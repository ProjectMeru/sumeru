package modelmeta_test

import (
	"testing"

	"sumeru/core/modelmeta"
)

func TestFieldNameFromGo(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"ID":         "id",
		"OrderLine":  "order_line",
		"OrderID":    "order_id",
		"MoveID":     "move_id",
		"LineIDs":    "line_ids",
		"LineIds":    "line_ids",
		"TagIDs":     "tag_ids",
		"TeamIDs":    "team_ids",
		"LeadIDs":    "lead_ids",
		"MemberIDs":  "member_ids",
		"CompanyIds": "company_ids",
	}
	for in, want := range cases {
		if got := modelmeta.FieldNameFromGo(in); got != want {
			t.Errorf("FieldNameFromGo(%q) = %q, want %q", in, got, want)
		}
	}
}
