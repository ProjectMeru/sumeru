package orm

import (
	"context"
)

// EnsureDefaultGroupsAndImplied creates kernel module categories, base security groups,
// xml ids base.group_system / base.group_user, and the Admin implies User edge.
// Call before module data sync (-i / -u) so addon XML can use ref('base.group_user') in implied_ids.
func EnsureDefaultGroupsAndImplied() error {
	ctx := AuditedBypass(context.Background(), "security.bootstrap")
	_, _, err := ensureDefaultKernelGroups(ctx)
	return err
}

// EnsureBootstrapSecurityFromSetup creates groups, the first administrator, company, and ACLs.
// Call only from /setup/init after base module install when the database has no users yet.
func EnsureBootstrapSecurityFromSetup(p SetupAdminParams) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return ensureBootstrapSecurity(context.Background(), &p)
}

// EnsureBootstrapSecurity ensures default groups and ACLs. If the database has no users yet,
// it returns an error so operators complete /setup instead of relying on a default account.
func EnsureBootstrapSecurity() error {
	return ensureBootstrapSecurity(context.Background(), nil)
}
