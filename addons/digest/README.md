# Digest addon

Scheduled email digests with KPI rows backed by `digest.digest` and `digest.kpi`.

## KPI compute codes

`ComputeKPI` in [`compute.go`](compute.go) evaluates each KPI’s `compute_code` at send time. Built-in codes today:

| Code | Meaning |
|------|---------|
| `crm.lead_count` | Count of `crm.lead` records |
| `crm.lead_won_count` | Won leads |
| `crm.opportunity_count` | Opportunities (`type = opportunity`) |

When a module needs a new metric, add a `case` in `ComputeKPI` (or a shared registry helper in this package) and reference the code from KPI records in XML/data. Unknown codes return an error so bad configuration surfaces in cron logs instead of silent zeroes.

If the target model is not installed, counts return `0` without error.
