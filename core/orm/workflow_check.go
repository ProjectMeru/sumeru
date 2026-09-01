package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// WorkflowTransitionInput describes a state transition check on a record.
type WorkflowTransitionInput struct {
	Model     string
	RecordID  int
	FromState string
	ToState   string
	UID       int
}

// CanWorkflowTransition reports whether uid may move record id on model from→to.
// Prefers sys.workflow.transition; falls back to sys.approval.rule.
func CanWorkflowTransition(ctx context.Context, in WorkflowTransitionInput) error {
	model := strings.TrimSpace(in.Model)
	toState := strings.TrimSpace(in.ToState)
	fromState := in.FromState
	uid := in.UID
	recordID := in.RecordID
	if model == "" || toState == "" {
		return fmt.Errorf("model and to_state required")
	}
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	if err := checkWorkflowTransitionRows(ctx, model, fromState, toState, uid); err != errNoWorkflowRows {
		return err
	}
	if recordID > 0 {
		return CheckStageApproval(ctx, model, recordID, toState)
	}
	return nil
}

var errNoWorkflowRows = errors.New("no workflow rows")

func checkWorkflowTransitionRows(ctx context.Context, model, fromState, toState string, uid int) error {
	if _, ok := Registry["sys.workflow.transition"]; !ok || DB == nil {
		return errNoWorkflowRows
	}
	tbl := MustQuotedTableName("sys.workflow.transition")
	rows, err := DB.QueryContext(ctx,
		`SELECT group_id, COALESCE(from_state,'') FROM `+tbl+
			` WHERE model = $1 AND to_state = $2 AND active = true`, model, toState)
	if err != nil {
		return errNoWorkflowRows
	}
	defer rows.Close()
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return err
	}
	n := 0
	matchedFrom := false
	allowed := false
	for rows.Next() {
		n++
		var gid sql.NullInt64
		var from string
		if err := rows.Scan(&gid, &from); err != nil {
			return err
		}
		if from != "" && from != fromState {
			continue
		}
		matchedFrom = true
		if !gid.Valid || gid.Int64 == 0 {
			allowed = true
			break
		}
		if intSliceContains(groups, int(gid.Int64)) {
			allowed = true
			break
		}
	}
	if n == 0 {
		return errNoWorkflowRows
	}
	if !matchedFrom || !allowed {
		return fmt.Errorf("workflow: transition %q → %q denied", fromState, toState)
	}
	return nil
}

// CheckStageApproval verifies if uid has permission to move record to targetState.
func CheckStageApproval(ctx context.Context, model string, id int, targetState string) error {
	if SecurityBypass(ctx) || SecurityUID(ctx) == superuserUID {
		return nil
	}
	uid := SecurityUID(ctx)
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return err
	}

	before, err := SearchOne(ctx, model, map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Errorf("approval denied: record lookup failed")
	}
	currentState := AsString(before["state"])

	appTbl := MustQuotedTableName("sys.approval.rule")
	// Find rules for this model and target state
	q := `SELECT group_id, COALESCE(from_state, '') FROM ` + appTbl + ` WHERE model = $1 AND to_state = $2 AND require_approval = true`
	rows, err := DB.QueryContext(ctx, q, model, targetState)
	if err != nil {
		// If table doesn't exist yet or other error, allow.
		return nil
	}
	defer rows.Close()

	hasRule := false
	match := false
	for rows.Next() {
		hasRule = true
		var gid int
		var fromState string
		if err := rows.Scan(&gid, &fromState); err != nil {
			return err
		}

		// If fromState is specified, it must match current state
		if fromState != "" && fromState != currentState {
			continue
		}

		if intSliceContains(groups, gid) {
			match = true
			break
		}
	}

	if hasRule && !match {
		return fmt.Errorf("approval required for transition to state %q (from %q)", targetState, currentState)
	}
	return nil
}
