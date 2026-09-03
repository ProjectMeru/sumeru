package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func testDateModel() orm.StubModelForTest {
	return orm.NewStubModelForTest("test.date", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char, Required: true, String: "Name"},
		{Name: "date_deadline", Type: orm.Date, String: "Deadline"},
		{Name: "date_last_stage_update", Type: orm.DateTime, String: "Last Stage Update"},
	})
}

func TestCoerceFieldValueEmptyDateBecomesNil(t *testing.T) {
	fd := orm.FieldDefinition{Name: "date_deadline", Type: orm.Date}
	got, err := orm.CoerceFieldValueForTest(fd, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for empty date, got %#v", got)
	}
}

func TestCoerceFieldValueEmptyDateTimeBecomesNil(t *testing.T) {
	fd := orm.FieldDefinition{Name: "date_last_stage_update", Type: orm.DateTime}
	got, err := orm.CoerceFieldValueForTest(fd, "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for empty datetime, got %#v", got)
	}
}

func TestCoerceFieldValueNonEmptyDatePreserved(t *testing.T) {
	fd := orm.FieldDefinition{Name: "date_deadline", Type: orm.Date}
	got, err := orm.CoerceFieldValueForTest(fd, "2026-08-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "2026-08-15" {
		t.Fatalf("expected date string preserved, got %#v", got)
	}
}

func TestPrepareValuesEmptyOptionalDateIsNil(t *testing.T) {
	out, err := orm.PrepareValues(testDateModel(), map[string]interface{}{
		"name":          "Acme",
		"date_deadline": "",
	}, orm.WriteOpCreate, orm.PrepareOptions{StrictUnknown: true})
	if err != nil {
		t.Fatalf("PrepareValues: %v", err)
	}
	if out["date_deadline"] != nil {
		t.Fatalf("expected nil date_deadline, got %#v", out["date_deadline"])
	}
}

func TestPrepareValuesRequiredFieldValidationError(t *testing.T) {
	_, err := orm.PrepareValues(testDateModel(), map[string]interface{}{}, orm.WriteOpCreate, orm.PrepareOptions{StrictUnknown: true})
	fve, ok := err.(*orm.FieldValidationError)
	if !ok {
		t.Fatalf("expected FieldValidationError, got %T: %v", err, err)
	}
	if fve.Field != "name" || fve.Label != "Name" {
		t.Fatalf("unexpected validation error: %+v", fve)
	}
}
