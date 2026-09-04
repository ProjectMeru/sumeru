package orm_test

import (
	"encoding/json"
	"testing"

	"sumeru/core/orm"
)

func TestCoerceFloat64(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   interface{}
		want float64
		ok   bool
	}{
		{float64(1.5), 1.5, true},
		{float32(2), 2, true},
		{int(3), 3, true},
		{int64(4), 4, true},
		{json.Number("2.5"), 2.5, true},
		{"3.25", 3.25, true},
		{[]byte("4.5"), 4.5, true},
		{"", 0, false},
		{true, 0, false},
	}
	for _, tc := range cases {
		got, ok := orm.CoerceFloat64(tc.in)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("CoerceFloat64(%v)=(%v,%v) want (%v,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
