package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func TestValidateAttachmentMIME_rejectsHTML(t *testing.T) {
	err := orm.ValidateAttachmentMIME("text/html", []byte("<html></html>"))
	if err == nil {
		t.Fatal("expected HTML upload to be rejected")
	}
}

func TestValidateAttachmentMIME_allowsPDF(t *testing.T) {
	if err := orm.ValidateAttachmentMIME("application/pdf", []byte("%PDF-1.4")); err != nil {
		t.Fatalf("expected PDF allowed: %v", err)
	}
}

func TestValidateAttachmentMIME_sniffsPlainText(t *testing.T) {
	if err := orm.ValidateAttachmentMIME("", []byte("hello,csv\n1,2")); err != nil {
		t.Fatalf("expected sniffed text/plain allowed: %v", err)
	}
}
