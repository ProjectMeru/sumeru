package web_test

import (
	"testing"

	"sumeru/core/server/web"
)

func TestFlashFromQueryImportMessage(t *testing.T) {
	flash, ok := web.FlashFromQueryMessage("imported_5_updated_2_skipped_1")
	if !ok {
		t.Fatal("expected flash")
	}
	if flash.Kind != "success" || flash.Title != "Import complete" {
		t.Fatalf("flash = %+v", flash)
	}
}

func TestFlashFromQueryLegacyImport(t *testing.T) {
	flash, ok := web.FlashFromQueryMessage("imported_3")
	if !ok || flash.Body != "Imported 3 row(s)." {
		t.Fatalf("flash = %+v ok=%v", flash, ok)
	}
}

func TestFlashFromQuerySaveError(t *testing.T) {
	flash, ok := web.FlashFromQueryMessage("save_error:invalid name")
	if !ok || flash.Kind != "error" || flash.Title != "Save failed" {
		t.Fatalf("flash = %+v ok=%v", flash, ok)
	}
}

func TestFlashFromQuerySaveOKCreated(t *testing.T) {
	flash, ok := web.FlashFromQueryMessage("save_ok_created")
	if !ok || flash.Kind != "success" || flash.ToastOnly {
		t.Fatalf("flash = %+v ok=%v", flash, ok)
	}
}

func TestFlashFromQuerySaveOKUpdated(t *testing.T) {
	flash, ok := web.FlashFromQueryMessage("save_ok_updated")
	if !ok || flash.Kind != "success" || flash.ToastOnly {
		t.Fatalf("flash = %+v ok=%v", flash, ok)
	}
}
