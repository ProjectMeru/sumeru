package sdk

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"sumeru/core/orm"
)

// dateLayout formats core.currency.rate.date_from values for domain filtering.
const dateLayout = "2006-01-02"

// ConvertCurrency converts amount expressed in fromCurrencyID into
// toCurrencyID as of date, using for each currency the newest active
// core.currency.rate whose date_from is <= date. A zero date means today.
//
// Rates express the value of one unit of the currency in the platform
// reference currency (the currency whose rate is 1.0), so the math is
// amount * rateFrom / rateTo.
func ConvertCurrency(ctx context.Context, amount float64, fromCurrencyID, toCurrencyID int, date time.Time) (float64, error) {
	if fromCurrencyID == toCurrencyID {
		return amount, nil
	}
	if date.IsZero() {
		date = time.Now()
	}
	rateFrom, err := currencyRate(ctx, fromCurrencyID, date)
	if err != nil {
		return 0, err
	}
	rateTo, err := currencyRate(ctx, toCurrencyID, date)
	if err != nil {
		return 0, err
	}
	return amount * rateFrom / rateTo, nil
}

// currencyRate resolves the newest effective rate for one currency.
func currencyRate(ctx context.Context, currencyID int, date time.Time) (float64, error) {
	domain := [][]interface{}{
		{"currency_id", "=", currencyID},
		{"date_from", "<=", date.Format(dateLayout)},
		{"active", "=", true},
	}
	rows, err := orm.SearchPage(ctx, "core.currency.rate", domain, 10, 0, "date_from desc")
	if err != nil {
		return 0, fmt.Errorf("sdk: lookup currency rate for currency %d: %w", currencyID, err)
	}
	if len(rows) == 0 {
		return 0, fmt.Errorf("sdk: no active core.currency.rate for currency %d on %s or earlier",
			currencyID, date.Format(dateLayout))
	}
	// Rows are ordered by date_from desc; on equal dates the highest id wins
	// so the result is deterministic.
	var best map[string]interface{}
	for _, row := range rows {
		if best == nil || rowKey(row) > rowKey(best) {
			best = row
		}
	}
	rate, ok := asFloat(best["rate"])
	if !ok {
		return 0, fmt.Errorf("sdk: non-numeric rate %v on core.currency.rate row %v", best["rate"], best["id"])
	}
	if rate <= 0 {
		return 0, fmt.Errorf("sdk: invalid non-positive rate %v for currency %d", rate, currencyID)
	}
	return rate, nil
}

// rowKey orders rate rows by (date_from, id) so ties on date_from resolve to
// the newest row.
func rowKey(row map[string]interface{}) string {
	id, _ := asInt(row["id"])
	return fmt.Sprintf("%s|%020d", dateString(row["date_from"]), id)
}

// ConvertUom converts a quantity from fromUomID into toUomID. Both units must
// belong to the same uom.category. The result is rounded to multiples of the
// target unit's rounding precision.
func ConvertUom(ctx context.Context, amount float64, fromUomID, toUomID int) (float64, error) {
	if fromUomID == toUomID {
		return amount, nil
	}
	rows, err := orm.Search(ctx, "uom.uom", [][]interface{}{{"id", "in", []interface{}{fromUomID, toUomID}}})
	if err != nil {
		return 0, fmt.Errorf("sdk: lookup uom %d and %d: %w", fromUomID, toUomID, err)
	}
	var from, to map[string]interface{}
	for _, row := range rows {
		id, _ := asInt(row["id"])
		switch id {
		case fromUomID:
			from = row
		case toUomID:
			to = row
		}
	}
	if from == nil || to == nil {
		return 0, fmt.Errorf("sdk: uom not found (from=%d to=%d)", fromUomID, toUomID)
	}
	fromCat, _ := asInt(from["category_id"])
	toCat, _ := asInt(to["category_id"])
	if fromCat != toCat {
		return 0, fmt.Errorf("sdk: cannot convert uom %d (category %d) to uom %d (category %d)",
			fromUomID, fromCat, toUomID, toCat)
	}
	factorFrom, ok := asFloat(from["factor"])
	if !ok || factorFrom <= 0 {
		return 0, fmt.Errorf("sdk: invalid factor on uom %d", fromUomID)
	}
	factorTo, ok := asFloat(to["factor"])
	if !ok || factorTo <= 0 {
		return 0, fmt.Errorf("sdk: invalid factor on uom %d", toUomID)
	}
	rounding, _ := asFloat(to["rounding"])
	return ConvertUomByFactor(amount, factorFrom, factorTo, rounding), nil
}

// ConvertUomByFactor applies the pure factor math of uom conversion:
// amount * factorFrom / factorTo, rounded to multiples of rounding (no-op
// when rounding <= 0). Exported so conversions can be unit-tested without a
// database and reused by verticals on preloaded records.
func ConvertUomByFactor(amount, factorFrom, factorTo, rounding float64) float64 {
	out := amount * factorFrom / factorTo
	if rounding > 0 {
		out = math.Round(out/rounding) * rounding
	}
	return out
}

// ConvertMoneyField converts the value of fieldName on rec (a record of
// modelName) into toCurrencyID as of date. fieldName must be tagged
// currency=<sibling many2one field>; the sibling field's value on rec names
// the source currency.
func ConvertMoneyField(ctx context.Context, modelName, fieldName string, rec map[string]interface{}, toCurrencyID int, date time.Time) (float64, error) {
	inst, ok := orm.Registry[modelName]
	if !ok {
		return 0, fmt.Errorf("sdk: unknown model %q", modelName)
	}
	var def orm.FieldDefinition
	found := false
	for _, f := range inst.Fields() {
		if f.Name == fieldName {
			def = f
			found = true
			break
		}
	}
	if !found {
		return 0, fmt.Errorf("sdk: model %s has no field %q", modelName, fieldName)
	}
	if def.Currency == "" {
		return 0, fmt.Errorf("sdk: field %s.%s is not tagged currency=", modelName, fieldName)
	}
	fromID, ok := asInt(rec[def.Currency])
	if !ok || fromID == 0 {
		return 0, fmt.Errorf("sdk: record has no %q value for %s.%s", def.Currency, modelName, fieldName)
	}
	amount, ok := asFloat(rec[fieldName])
	if !ok {
		return 0, fmt.Errorf("sdk: record has no numeric value for %s.%s", modelName, fieldName)
	}
	return ConvertCurrency(ctx, amount, fromID, toCurrencyID, date)
}

// dateString renders scanned DATE values (time.Time or driver string) back to
// the ISO date layout used for lexicographic comparison.
func dateString(v interface{}) string {
	switch t := v.(type) {
	case time.Time:
		return t.Format(dateLayout)
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

func asInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		return i, err == nil
	case []byte:
		i, err := strconv.Atoi(strings.TrimSpace(string(n)))
		return i, err == nil
	default:
		return 0, false
	}
}

func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f, err == nil
	case []byte:
		f, err := strconv.ParseFloat(strings.TrimSpace(string(n)), 64)
		return f, err == nil
	default:
		return 0, false
	}
}
