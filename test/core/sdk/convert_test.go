package sdk_test

import (
	"context"
	"math"
	"testing"
	"time"

	"sumeru/core/orm"
	"sumeru/core/sdk"
)

func TestConvertUomByFactorMath(t *testing.T) {
	cases := []struct {
		name                       string
		amount, from, to, rounding float64
		want                       float64
	}{
		{"kg to g", 1, 1, 0.001, 0.001, 1000},
		{"g to kg", 500, 0.001, 1, 0.001, 0.5},
		{"dozen to unit", 2, 12, 1, 1, 24},
		{"unit to dozen rounding", 13, 1, 12, 1, 1},
		{"no rounding when zero", 1.0 / 3, 1, 1, 0, 1.0 / 3},
		{"round to two decimals", 1, 1, 3, 0.01, 0.33},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sdk.ConvertUomByFactor(tc.amount, tc.from, tc.to, tc.rounding)
			if math.Abs(got-tc.want) > 1e-9 {
				t.Fatalf("ConvertUomByFactor(%v,%v,%v,%v) = %v, want %v",
					tc.amount, tc.from, tc.to, tc.rounding, got, tc.want)
			}
		})
	}
}

// Fast paths must not touch the database.
func TestConvertCurrencySameCurrencyNoDB(t *testing.T) {
	ctx := context.Background()
	got, err := sdk.ConvertCurrency(ctx, 123.45, 7, 7, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 123.45 {
		t.Fatalf("got %v, want 123.45", got)
	}
}

func TestConvertUomSameUnitNoDB(t *testing.T) {
	ctx := context.Background()
	got, err := sdk.ConvertUom(ctx, 42, 3, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Fatalf("got %v, want 42", got)
	}
}

func TestConvertMoneyFieldUnknownModel(t *testing.T) {
	_, err := sdk.ConvertMoneyField(context.Background(), "no.such.model", "amount", nil, 1, time.Time{})
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
}

func TestConvertMoneyFieldRequiresCurrencyTag(t *testing.T) {
	orm.RegisterStubModelForTest(t, "stub.plain", []orm.FieldDefinition{
		{Name: "amount", Type: orm.Float64},
	})
	_, err := sdk.ConvertMoneyField(context.Background(), "stub.plain", "amount", map[string]interface{}{"amount": 10}, 1, time.Time{})
	if err == nil {
		t.Fatal("expected error for field without currency tag")
	}
}

func TestConvertMoneyFieldRequiresSiblingValue(t *testing.T) {
	orm.RegisterStubModelForTest(t, "stub.money", []orm.FieldDefinition{
		{Name: "amount", Type: orm.Float64, Currency: "currency_id"},
	})
	rec := map[string]interface{}{"amount": 10}
	_, err := sdk.ConvertMoneyField(context.Background(), "stub.money", "amount", rec, 1, time.Time{})
	if err == nil {
		t.Fatal("expected error for missing sibling currency value")
	}
}

func TestConvertMoneyFieldRequiresNumericAmount(t *testing.T) {
	orm.RegisterStubModelForTest(t, "stub.money2", []orm.FieldDefinition{
		{Name: "amount", Type: orm.Float64, Currency: "currency_id"},
	})
	rec := map[string]interface{}{"amount": "not-a-number", "currency_id": 7}
	_, err := sdk.ConvertMoneyField(context.Background(), "stub.money2", "amount", rec, 1, time.Time{})
	if err == nil {
		t.Fatal("expected error for non-numeric amount")
	}
}
