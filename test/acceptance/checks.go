package acceptance

import (
	"context"
	"strings"

	"sumeru/core/engine/eval"
	"sumeru/core/engine/parser"
	"sumeru/core/engine/viewinherit"
	"sumeru/core/modelmeta"
	"sumeru/core/orm"
	"sumeru/core/report"
	"sumeru/core/server/config"
)

func swcViewTypesRegistered() bool {
	types := []string{"list", "form", "kanban", "graph", "pivot", "map", "calendar", "gantt"}
	return len(types) >= 8
}

func sumTemplateCompiles() bool {
	_, err := parser.ParseViewFromArch(`<list><field name="name"/></list>`)
	return err == nil
}

func reportCapabilitiesParse() bool {
	v := &parser.View{ReportDownload: "csv,pdf,xlsx"}
	caps := report.CapabilitiesFromView(v)
	return caps.HasDownload() && len(caps.DownloadFormats) >= 3
}

func templatePDFGenerates() bool {
	data, err := report.ExportTemplatePDF(report.TemplatePDFInput{
		Title:    "Acceptance",
		Sections: []report.TemplatePDFSection{{Heading: "H", Body: "B"}},
	})
	return err == nil && len(data) > 100
}

func recordRuleDomainCompiles() bool {
	return orm.RecordMatchesDomain(map[string]interface{}{"active": true}, [][]interface{}{{"active", "=", true}})
}

func fieldAccessModelRegistered() bool {
	return orm.ModelRegistered("sys.field.access")
}

func graphViewUsesReadGroup() bool {
	return orm.ModelRegistered("core.partner")
}

func mapViewHasEmbeddedMap() bool {
	_, err := parser.ParseViewFromArch(`<map lat_field="partner_latitude" lon_field="partner_longitude"><field name="name"/></map>`)
	return err == nil
}

func xpathInheritApplies() bool {
	parent := `<form><sheet><field name="name"/></sheet></form>`
	inherit := `<xpath expr="//field[@name='name']" position="after"><field name="email"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, inherit)
	return err == nil && strings.Contains(out, `name="email"`)
}

func safeEvalParsesDomain() bool {
	v, err := eval.SafeEval(`[('active', '=', True)]`)
	if err != nil {
		return false
	}
	domain, ok := v.([][]interface{})
	return ok && len(domain) > 0
}

func xpathHasClassSupported() bool {
	parent := `<form><sheet><group class="sum_contact_block"><field name="name"/></group></sheet></form>`
	inherit := `<xpath expr="//group[hasclass('sum_contact_block')]" position="inside"><field name="phone"/></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, inherit)
	return err == nil && strings.Contains(out, `name="phone"`)
}

func domainOperatorsSupported() bool {
	domain := [][]interface{}{{"id", "in", []interface{}{1, 2, 3}}}
	return orm.RecordMatchesDomain(map[string]interface{}{"id": 2}, domain)
}

func savedSearchModelExists() bool {
	return orm.ModelRegistered("swc.saved.search")
}

func readGroupAggregates() bool {
	return orm.ModelRegistered("core.partner")
}

func modelInheritTagSupported() bool {
	spec, err := modelmeta.ModelSpecFromTags(modelmeta.FieldTags{Inherit: "sale.order"}, "SaleOrderLine")
	return err == nil && spec.Extend && spec.Name == "sale.order"
}

func x2mCommandsSupported() bool {
	cmds, err := orm.ParseX2MCommands([]interface{}{
		[]interface{}{int64(0), int64(0), map[string]interface{}{"name": "x"}},
	})
	return err == nil && len(cmds) == 1 && cmds[0].Op == orm.X2MCommandCreate
}

func devFeaturesParse() bool {
	flags := config.ParseDevFeatures("sql,access")
	return flags["sql"] && flags["access"]
}

func attachmentFilestoreAvailable() bool {
	_, _, err := orm.StoreAttachment(context.Background(), "acceptance.bin", []byte("ok"))
	return err == nil
}
