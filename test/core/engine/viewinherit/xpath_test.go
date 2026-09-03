package viewinherit_test

import (
	"strings"
	"testing"

	"sumeru/core/engine/viewinherit"
)

func TestApplyInheritArchHasClass(t *testing.T) {
	parent := `<form><sheet><group class="sum_contact_block"><field name="name"/></group></sheet></form>`
	frag := `<xpath expr="//group[hasclass('sum_contact_block')]" position="inside"><field name="phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `name="phone"`) {
		t.Fatalf("hasclass xpath failed: %s", out)
	}
}

func TestApplyInheritArchAfter(t *testing.T) {
	parent := `<view model="sale.order" type="list"><field name="a" string="A"/><field name="b" string="B"/></view>`
	frag := `<xpath expr="//field[@name='a']" position="after"><field name="z" string="Z"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `name="a"`, `name="z"`, `name="b"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchAfterMarshaledField(t *testing.T) {
	parent := `<view id="" model="sale.order" type="list" title="" open=""><field name="state" string="Status" widget="" placeholder="" options=""></field><field name="amount" string="Total" widget="" placeholder="" options=""></field></view>`
	frag := `<xpath expr="//field[@name='state']" position="after"><field name="phone" string="Phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `name="state"`, `name="phone"`, `name="amount"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchButtonAndAttributes(t *testing.T) {
	parent := `<view type="form"><header><button name="action_lost" string="Lost" type="object"></button></header><field name="phone" string="Phone"></field></view>`
	frag := `<xpath expr="//button[@name='action_lost']" position="attributes"><attribute name="string">Mark Lost</attribute></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `string="Mark Lost"`) {
		t.Fatalf("button attr: %s", out)
	}
	frag2 := `<xpath expr="//field[@name='phone']" position="attributes"><attribute name="invisible">1</attribute></xpath>`
	out2, err := viewinherit.ApplyInheritArch(out, frag2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out2, `invisible="1"`) {
		t.Fatalf("field attr: %s", out2)
	}
}

func TestApplyInheritArchGroupInside(t *testing.T) {
	parent := `<view type="form"><sheet><group string="Identity"><field name="name"/></group></sheet></view>`
	frag := `<xpath expr="//group[@string='Identity']" position="inside"><field name="phone" string="Phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `string="Identity"`, `name="name"`, `name="phone"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchNestedGroupInside(t *testing.T) {
	parent := `<view type="form"><group string="Outer"><group string="Inner"><field name="name"/></group></group></view>`
	frag := `<xpath expr="//group[@string='Outer']" position="inside"><field name="phone" string="Phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `string="Outer"`, `string="Inner"`, `name="name"`, `name="phone"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchSingleQuotedAttr(t *testing.T) {
	parent := `<view type="form"><field name='phone' string="Phone"/></view>`
	frag := `<xpath expr="//field[@name='phone']" position="attributes"><attribute name="invisible">1</attribute></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `invisible="1"`) {
		t.Fatalf("field attr: %s", out)
	}
}

func TestApplyInheritArchSingleQuotedAttrReplace(t *testing.T) {
	parent := `<view type="form"><field name='phone' string='Old'/></view>`
	frag := `<xpath expr="//field[@name='phone']" position="attributes"><attribute name="string">New</attribute></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `string="New"`) {
		t.Fatalf("field attr: %s", out)
	}
	if strings.Contains(out, `string='Old'`) || strings.Contains(out, `string='New'`) {
		t.Fatalf("expected double-quoted replacement, got: %s", out)
	}
	if strings.Count(out, `name='phone'`) != 1 && strings.Count(out, `name="phone"`) != 1 {
		t.Fatalf("duplicate name attr: %s", out)
	}
	if strings.Count(out, "name=") > 1 {
		t.Fatalf("duplicate name attr: %s", out)
	}
}

func TestApplyInheritArchGroupAfter(t *testing.T) {
	parent := `<view type="form"><group string="Outer"><group string="Inner"><field name="name"/></group></group></view>`
	frag := `<xpath expr="//group[@string='Outer']" position="after"><field name="phone" string="Phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `string="Inner"`, `name="name"`, `</group>`, `name="phone"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchGroupBefore(t *testing.T) {
	parent := `<view type="form"><group string="Outer"><field name="name"/></group></view>`
	frag := `<xpath expr="//group[@string='Outer']" position="before"><field name="phone" string="Phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !containsInOrder(out, `name="phone"`, `string="Outer"`, `name="name"`) {
		t.Fatalf("unexpected merge: %s", out)
	}
}

func TestApplyInheritArchSheetReplace(t *testing.T) {
	parent := `<view type="form"><sheet><field name="name"/></sheet></view>`
	frag := `<xpath expr="//sheet" position="replace"><sheet><field name="replaced"/></sheet></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, `name="name"`) {
		t.Fatalf("expected original field removed: %s", out)
	}
	if !strings.Contains(out, `name="replaced"`) {
		t.Fatalf("expected replacement field: %s", out)
	}
}

func TestApplyInheritArchFirstMatchWins(t *testing.T) {
	parent := `<view type="form"><field name="phone" string="First"/><field name="phone" string="Second"/></view>`
	frag := `<xpath expr="//field[@name='phone']" position="attributes"><attribute name="string">Updated</attribute></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	firstIdx := strings.Index(out, `string="Updated"`)
	secondIdx := strings.Index(out, `string="Second"`)
	if firstIdx < 0 || secondIdx < 0 {
		t.Fatalf("unexpected merge: %s", out)
	}
	if firstIdx > secondIdx {
		t.Fatalf("expected first field updated only: %s", out)
	}
	if strings.Count(out, `string="Updated"`) != 1 {
		t.Fatalf("expected single update: %s", out)
	}
}

func containsInOrder(s string, parts ...string) bool {
	pos := 0
	for _, p := range parts {
		i := indexFrom(s, p, pos)
		if i < 0 {
			return false
		}
		pos = i + len(p)
	}
	return true
}

func indexFrom(s, sub string, start int) int {
	if start > len(s) {
		return -1
	}
	idx := strings.Index(s[start:], sub)
	if idx < 0 {
		return -1
	}
	return start + idx
}
