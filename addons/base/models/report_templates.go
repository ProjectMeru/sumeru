package models

import "sumeru/core/report"

func init() {
	report.RegisterHTMLTemplate("demo_platform_report", `<h1>{{.Report.name}}</h1>
<p>Model: {{.Report.model}}</p>
{{if .Record.name}}<p>Record: {{.Record.name}}</p>{{end}}
<table>
<tr><th>Field</th><th>Value</th></tr>
<tr><td>Report</td><td>{{.Report.name}}</td></tr>
</table>`)
}
