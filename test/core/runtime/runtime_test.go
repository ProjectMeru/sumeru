package runtime_test

import (
	"context"
	"testing"

	"sumeru/core/event"
	"sumeru/core/orm"
	"sumeru/core/runtime"
)

type stubModel struct{}

func (stubModel) ModelName() string { return "test.runtime" }
func (stubModel) Fields() []orm.FieldDefinition {
	return []orm.FieldDefinition{{Name: "name", Type: orm.Char}}
}

func TestRuntime_isolatedRegistry(t *testing.T) {
	rt := runtime.New(nil, nil)
	if _, ok := rt.Model("test.runtime"); ok {
		t.Fatal("expected empty registry")
	}
	rt.RegisterModel(stubModel{})
	m, ok := rt.Model("test.runtime")
	if !ok || m.ModelName() != "test.runtime" {
		t.Fatalf("got %#v ok=%v", m, ok)
	}
}

func TestRuntime_SetDefault(t *testing.T) {
	runtime.SetDefault(nil)
	defer runtime.SetDefault(nil)

	rt := runtime.New(nil, map[string]orm.Model{"test.runtime": stubModel{}})
	runtime.SetDefault(rt)
	got := runtime.Default()
	if got != rt {
		t.Fatal("SetDefault did not stick")
	}
	if _, ok := got.Model("test.runtime"); !ok {
		t.Fatal("model missing on default")
	}
}

func TestRuntime_SyncFromGlobalsAndPublish(t *testing.T) {
	runtime.SetDefault(nil)
	defer runtime.SetDefault(nil)

	runtime.SyncFromGlobals()
	rt := runtime.Default()
	if rt == nil || rt.Events == nil {
		t.Fatal("default runtime")
	}
	var nilRT *runtime.Runtime
	nilRT.RegisterModel(stubModel{})
	if _, ok := nilRT.Model("test.runtime"); ok {
		t.Fatal("nil runtime register")
	}
}

func TestEventBusPublish(t *testing.T) {
	bus := runtime.New(nil, nil).Events
	errs := bus.Publish(context.Background(), event.Event{Name: "noop.test"})
	if len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
}
