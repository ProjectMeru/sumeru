package orm

// superuserUID is break-glass kernel access (uid=1). New code should prefer group_system
// or AuditedBypass/WithElevated over numeric uid checks.
const superuserUID = 1

const (
	tableGroupUserRel   = "core.group.user.rel"
	tableGroupImplied   = "core.group.implied"
	tableRuleGroupRel   = "sys.rule.group.rel"
	tableUserCompanyRel = "core.user.company.rel"
)
