package orm_test

import (
	"context"
	"fmt"
	"testing"
	"sumeru/core/orm"
)


func TestApplyComputes(t *testing.T) {
	orm.RegisterCompute("test.compute", "full_name", []string{"first", "last"}, func(_ context.Context, rec map[string]interface{}) (interface{}, error) {
		return orm.AsString(rec["first"]) + " " + orm.AsString(rec["last"]), nil
	})
	rec := map[string]interface{}{"first": "Ada", "last": "Lovelace"}
	if err := orm.ApplyComputes(context.Background(), "test.compute", rec); err != nil {
		t.Fatal(err)
	}
	if rec["full_name"] != "Ada Lovelace" {
		t.Fatalf("got %q", rec["full_name"])
	}
}

func TestMergeDomainsAND(t *testing.T) {
	base := [][]interface{}{{"active", "=", true}}
	search := [][]interface{}{{"name", "ilike", "%acme%"}}
	out := orm.MergeDomains(base, search)
	if len(out) != 2 {
		t.Fatalf("len=%d", len(out))
	}
}

func TestBuildListSearchDomainOR(t *testing.T) {
	dom := orm.BuildListSearchDomain("sys.module", []string{"name", "state"}, "base")
	if len(dom) < 3 {
		t.Fatalf("expected OR domain, got %v", dom)
	}
	if fmt.Sprint(dom[0][0]) != "|" {
		t.Fatalf("expected leading OR, got %v", dom)
	}
}
