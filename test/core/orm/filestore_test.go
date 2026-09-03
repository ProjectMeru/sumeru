package orm_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"sumeru/core/orm"
)

func TestFilestoreRoundTrip(t *testing.T) {
	root := filepath.Join(t.TempDir(), "filestore")
	if err := orm.InitFilestore(root); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	data := []byte("attachment-bytes")
	key, size, err := orm.StoreAttachment(ctx, "doc.pdf", data)
	if err != nil || size != int64(len(data)) || key != "doc.pdf" {
		t.Fatalf("StoreAttachment: key=%q size=%d err=%v", key, size, err)
	}
	got, err := orm.ReadAttachment(ctx, key)
	if err != nil || string(got) != string(data) {
		t.Fatalf("ReadAttachment: %q err=%v", got, err)
	}
	rc, err := orm.OpenAttachment(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	all, err := io.ReadAll(rc)
	if err != nil || string(all) != string(data) {
		t.Fatalf("OpenAttachment: %q err=%v", all, err)
	}
}

func TestFilestoreErrors(t *testing.T) {
	ctx := context.Background()
	if _, _, err := orm.StoreAttachment(ctx, "x", nil); err == nil {
		t.Fatal("empty data")
	}
	if err := orm.InitFilestore(""); err != nil {
		t.Fatal(err)
	}
	key, _, err := orm.StoreAttachment(ctx, "", []byte("hash-key"))
	if err != nil || key == "" {
		t.Fatalf("anonymous key: %q err=%v", key, err)
	}
	if _, err := orm.ReadAttachment(ctx, "missing-"+t.Name()); err == nil {
		t.Fatal("missing file")
	}
	_ = os.RemoveAll(filepath.Join(os.TempDir(), "sumeru-filestore"))
}
