package render_test

import (
	"strings"
	"testing"
	"sumeru/core/engine/render"
)


func TestParseImageCrop_defaults(t *testing.T) {
	c, ok := render.ParseImageCrop("")
	if ok {
		t.Fatal("expected inactive for empty crop")
	}
	if c.X != 50 || c.Y != 50 || c.Zoom != 1 {
		t.Fatalf("unexpected defaults: %+v", c)
	}
}

func TestParseImageCrop_valid(t *testing.T) {
	c, ok := render.ParseImageCrop(`{"x":30,"y":40,"zoom":1.5}`)
	if !ok {
		t.Fatal("expected active crop")
	}
	if c.X != 30 || c.Y != 40 || c.Zoom != 1.5 {
		t.Fatalf("unexpected crop: %+v", c)
	}
}

func TestAvatarCropStyle(t *testing.T) {
	attr := render.AvatarCropStyle(render.ImageCrop{X: 25, Y: 75, Zoom: 1.2}, true)
	s := string(attr)
	if !strings.Contains(s, "object-position:25.00% 75.00%") {
		t.Fatalf("missing object-position: %q", s)
	}
	if !strings.Contains(s, "scale(1.200)") {
		t.Fatalf("missing scale: %q", s)
	}
}
