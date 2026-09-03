package viewinherit_test

import (
	"strings"
	"testing"

	"sumeru/core/engine/viewinherit"
)

func TestApplyInheritArch_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		parent string
		frag   string
		want   []string
		err    bool
	}{
		{
			name:   "tag only inside",
			parent: `<form><sheet><field name="x"/></sheet></form>`,
			frag:   `<xpath expr="//sheet" position="inside"><field name="note"/></xpath>`,
			want:   []string{`name="note"`, `name="x"`},
		},
		{
			name:   "replace header",
			parent: `<view type="form"><header><button name="old"/></header></view>`,
			frag:   `<xpath expr="//header" position="replace"><header><button name="new" string="New"/></header></xpath>`,
			want:   []string{`name="new"`, `string="New"`},
		},
		{
			name:   "before field",
			parent: `<view type="list"><field name="a"/><field name="b"/></view>`,
			frag:   `<xpath expr="//field[@name='b']" position="before"><field name="z"/></xpath>`,
			want:   []string{`name="a"`, `name="z"`, `name="b"`},
		},
		{
			name:   "data wrapper stripped",
			parent: `<form><field name="x"/></form>`,
			frag:   `<data><xpath expr="//field[@name='x']" position="attributes"><attribute name="readonly">1</attribute></xpath></data>`,
			want:   []string{`readonly="1"`},
		},
		{
			name:   "invalid xpath",
			parent: `<form><field name="x"/></form>`,
			frag:   `<xpath expr="//missing[@name='nope']" position="inside"><field name="y"/></xpath>`,
			err:    true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := viewinherit.ApplyInheritArch(tc.parent, tc.frag)
			if tc.err {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, part := range tc.want {
				if !strings.Contains(out, part) {
					t.Fatalf("output missing %q:\n%s", part, out)
				}
			}
		})
	}
}

func TestApplyInheritArch_orderPreserved(t *testing.T) {
	parent := `<view type="list"><field name="a"/><field name="b"/></view>`
	frag := `<xpath expr="//field[@name='a']" position="after"><field name="mid"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `name="a"`, `name="mid"`, `name="b"`) {
		t.Fatalf("order: %s", out)
	}
}
