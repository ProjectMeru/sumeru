package sdk_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/sdk"
)

type computeTotalModel struct {
	sdk.Model `sumeru:"model=test.sdk.compute.total"`
	Total     sdk.Numeric `sumeru:"string=Total"`
}

func TestBindComputeCoercesFieldType(t *testing.T) {
	modelreg.ResetActivationForTest()
	orm.Registry = map[string]orm.Model{}
	modelreg.MustRegister("test", &computeTotalModel{})
	if err := modelreg.ActivateAll(nil); err != nil {
		t.Fatal(err)
	}
	modelName := "test.sdk.compute.total"
	sdk.BindCompute(modelName, "total", func(context.Context, map[string]interface{}) (any, error) {
		return 12.5, nil
	})
	rec := map[string]interface{}{}
	if err := orm.ApplyComputes(context.Background(), modelName, rec); err != nil {
		t.Fatal(err)
	}
	if rec["total"] != 12.5 {
		t.Fatalf("got %#v", rec["total"])
	}

	sdk.BindCompute(modelName, "total", func(context.Context, map[string]interface{}) (any, error) {
		return "bad", nil
	})
	err := orm.ApplyComputes(context.Background(), modelName, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "total") {
		t.Fatalf("got %v", err)
	}
}

type statusHookModel struct {
	sdk.Model `sumeru:"model=test.sdk.status.hook"`
	Stage     sdk.Selection[string] `sumeru:"status,selection=draft:Draft,planned:Planned"`
}

func TestBindStatusStoresSelection(t *testing.T) {
	modelreg.ResetActivationForTest()
	orm.Registry = map[string]orm.Model{}
	modelreg.MustRegister("test", &statusHookModel{})
	if err := modelreg.ActivateAll(nil); err != nil {
		t.Fatal(err)
	}
	modelName := "test.sdk.status.hook"
	sdk.BindStatus(modelName, "stage", func(context.Context, map[string]interface{}) (any, error) {
		return "planned", nil
	})
	rec := map[string]interface{}{"stage": "draft"}
	if err := orm.ApplyStatus(context.Background(), modelName, "stage", rec); err != nil {
		t.Fatal(err)
	}
	if rec["stage"] != "planned" {
		t.Fatalf("got %#v", rec["stage"])
	}

	sdk.BindStatus(modelName, "stage", func(context.Context, map[string]interface{}) (any, error) {
		return "nope", nil
	})
	err := orm.ApplyStatus(context.Background(), modelName, "stage", map[string]interface{}{"stage": "draft"})
	if err == nil || !strings.Contains(err.Error(), "stage") {
		t.Fatalf("got %v", err)
	}
}
