package mail_test

import (
	"context"
	"testing"

	coremail "sumeru/core/mail"
)

func TestConfigureAndConfigured(t *testing.T) {
	coremail.Configure(coremail.SMTPConfig{})
	if coremail.Configured() {
		t.Fatal("empty config not configured")
	}
	coremail.Configure(coremail.SMTPConfig{Host: "smtp.example.com", From: "noreply@example.com"})
	if !coremail.Configured() {
		t.Fatal("expected configured")
	}
	coremail.Configure(coremail.SMTPConfig{Host: "smtp.example.com", Port: 0, From: "noreply@example.com"})
	if err := coremail.Send(context.Background(), "", "subj", "body"); err == nil {
		t.Fatal("empty recipient")
	}
	if err := coremail.Send(context.Background(), "user@example.com", "subj", "body"); err == nil {
		t.Fatal("Send without server should fail")
	}
}

func TestSendPasswordResetEmailNotConfigured(t *testing.T) {
	coremail.Configure(coremail.SMTPConfig{})
	if err := coremail.SendPasswordResetEmail(context.Background(), "u@x.com", "admin", "http://localhost/"); err == nil {
		t.Fatal("expected smtp not configured error")
	}
}

func TestEnqueueDoesNotPanic(t *testing.T) {
	coremail.Enqueue(context.Background(), "nobody@example.com", "subj", "body")
}
