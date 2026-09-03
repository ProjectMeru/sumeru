package mail_test

import (
	"context"
	"testing"

	"sumeru/addons/mail"
)

func TestPostMessage_requiresFields(t *testing.T) {
	err := mail.PostMessage(context.Background(), "", 1, "hi", mail.SubtypeComment, "A")
	if err == nil {
		t.Fatal("expected error for empty model")
	}
	err = mail.PostMessage(context.Background(), "core.user", 1, "", mail.SubtypeComment, "A")
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	err = mail.PostMessage(context.Background(), "core.user", 1, "hi", "", "A")
	if err == nil {
		t.Fatal("expected error for empty subtype")
	}
}

func TestPostMessage_unknownModelWithoutDB(t *testing.T) {
	// Without DB init, still validates model registry membership.
	err := mail.PostMessage(context.Background(), "no.such.model", 1, "body", mail.SubtypeComment, "A")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompanyFlags_defaultTrueWithoutDB(t *testing.T) {
	ctx := context.Background()
	if !mail.CompanyChatterEnabled(ctx) {
		t.Fatal("default chatter enabled")
	}
	if !mail.CompanyActivityPanelEnabled(ctx) {
		t.Fatal("default activity panel enabled")
	}
}

func TestListComments_nilDB(t *testing.T) {
	rows, err := mail.ListCommentsForRecord(context.Background(), "core.user", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if rows != nil && len(rows) != 0 {
		t.Fatalf("expected empty, got %v", rows)
	}
}
