package orm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"sumeru/core/orm"
)

func TestWithElevatedEmptyReason(t *testing.T) {
	err := orm.WithElevated(context.Background(), "  ", func(context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error for empty reason")
	}
}

func TestWithElevatedBypassAndDeadline(t *testing.T) {
	parent := context.Background()
	start := time.Now()
	var gotBypass bool
	var deadline time.Time
	err := orm.WithElevated(parent, "test.elevate", func(ctx context.Context) error {
		gotBypass = orm.BypassFromContext(ctx)
		d, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected deadline on elevated context")
		}
		deadline = d
		return nil
	})
	if err != nil {
		t.Fatalf("WithElevated: %v", err)
	}
	if !gotBypass {
		t.Fatal("expected bypass in callback")
	}
	maxEnd := start.Add(15*time.Minute + time.Second)
	if deadline.After(maxEnd) {
		t.Fatalf("deadline too far: %v", deadline)
	}
}

func TestWithElevatedPropagatesError(t *testing.T) {
	want := errors.New("boom")
	err := orm.WithElevated(context.Background(), "test.fail", func(context.Context) error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("got %v want %v", err, want)
	}
}

func TestWithElevatedNested(t *testing.T) {
	calls := 0
	err := orm.WithElevated(context.Background(), "outer", func(outer context.Context) error {
		calls++
		return orm.WithElevated(outer, "inner", func(inner context.Context) error {
			calls++
			if !orm.BypassFromContext(inner) {
				t.Fatal("inner expected bypass")
			}
			return nil
		})
	})
	if err != nil {
		t.Fatalf("nested: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestRejectSmuggledUserBypass_anonymousOK(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	if err := orm.RejectSmuggledUserBypass(ctx); err != nil {
		t.Fatalf("uid 0 bypass: %v", err)
	}
}

func TestAuditedBypassEmptyReason(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 1)
	out := orm.AuditedBypass(ctx, "")
	if orm.BypassFromContext(out) {
		t.Fatal("empty reason must not enable bypass")
	}
}
