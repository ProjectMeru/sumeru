package orm

import (
	"fmt"
	"strings"
)

type falsyKind int

const (
	falsyNone falsyKind = iota
	falsyNull
	falsyEmpty
	falsyNullOrZero
	falsyRelOne2Many
	falsyRelMany2Many
)

func falsyKindForField(fieldDef FieldDefinition) falsyKind {
	switch fieldDef.Type {
	case Date, DateTime, Many2One, Json:
		return falsyNull
	case Char, Text, Selection:
		return falsyEmpty
	case Integer, Float, Float64, Numeric:
		return falsyNullOrZero
	case One2Many:
		return falsyRelOne2Many
	case Many2Many:
		return falsyRelMany2Many
	default:
		return falsyNone
	}
}

func falsyWantSet(op string, value bool) (bool, bool) {
	switch strings.TrimSpace(strings.ToLower(op)) {
	case "=":
		return value, true
	case "!=":
		return !value, true
	default:
		return false, false
	}
}

func lookupFalsyField(modelName, fieldName string) (FieldDefinition, falsyKind, bool) {
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return FieldDefinition{}, falsyNone, false
	}
	for _, fieldDef := range inst.Fields() {
		if fieldDef.Name != fieldName {
			continue
		}
		kind := falsyKindForField(fieldDef)
		if kind == falsyNone {
			return fieldDef, falsyNone, false
		}
		return fieldDef, kind, true
	}
	return FieldDefinition{}, falsyNone, false
}

func buildFalsyDomainClause(modelName, fieldName, op string, value interface{}) (clause string, args []interface{}, handled bool, err error) {
	b, ok := value.(bool)
	if !ok {
		return "", nil, false, nil
	}
	wantSet, ok := falsyWantSet(op, b)
	if !ok {
		return "", nil, false, nil
	}
	fieldDef, kind, ok := lookupFalsyField(modelName, fieldName)
	if !ok {
		return "", nil, false, nil
	}
	switch kind {
	case falsyNull, falsyEmpty, falsyNullOrZero:
		col, err := QuotedColumnForModel(modelName, fieldName)
		if err != nil {
			return "", nil, false, err
		}
		return falsyColumnClause(col, kind, wantSet), nil, true, nil
	case falsyRelOne2Many:
		clause, err := falsyOne2ManyClause(modelName, fieldDef, wantSet)
		return clause, nil, true, err
	case falsyRelMany2Many:
		clause, err := falsyMany2ManyClause(modelName, fieldDef, wantSet)
		return clause, nil, true, err
	default:
		return "", nil, false, nil
	}
}

func falsyColumnClause(col string, kind falsyKind, wantSet bool) string {
	switch kind {
	case falsyNull:
		if wantSet {
			return col + " IS NOT NULL"
		}
		return col + " IS NULL"
	case falsyEmpty:
		if wantSet {
			return fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", col, col)
		}
		return fmt.Sprintf("(%s IS NULL OR %s = '')", col, col)
	case falsyNullOrZero:
		if wantSet {
			return fmt.Sprintf("(%s IS NOT NULL AND %s <> 0)", col, col)
		}
		return fmt.Sprintf("(%s IS NULL OR %s = 0)", col, col)
	default:
		return "TRUE"
	}
}

func falsyOne2ManyClause(parentModel string, fieldDef FieldDefinition, wantSet bool) (string, error) {
	comodel := strings.TrimSpace(fieldDef.Relation)
	if comodel == "" {
		return "", fmt.Errorf("one2many field %q on %s missing relation", fieldDef.Name, parentModel)
	}
	inverse := ResolveInverseOne2ManyField(parentModel, comodel)
	if inverse == "" {
		return "", fmt.Errorf("one2many field %q on %s: no inverse many2one on %s", fieldDef.Name, parentModel, comodel)
	}
	parentTbl, err := QuotedTableName(parentModel)
	if err != nil {
		return "", err
	}
	childTbl, err := QuotedTableForModel(comodel)
	if err != nil {
		return "", err
	}
	inverseCol, err := QuotedColumnForModel(comodel, inverse)
	if err != nil {
		return "", err
	}
	parentIDCol := parentTbl + "." + quoteIdent("id")
	exists := fmt.Sprintf("EXISTS (SELECT 1 FROM %s rel WHERE rel.%s = %s)", childTbl, inverseCol, parentIDCol)
	if wantSet {
		return exists, nil
	}
	return "NOT " + exists, nil
}

func falsyMany2ManyClause(parentModel string, fieldDef FieldDefinition, wantSet bool) (string, error) {
	relTable := strings.TrimSpace(fieldDef.RelationTable)
	leftCol := strings.TrimSpace(fieldDef.Column1)
	if relTable == "" || leftCol == "" {
		return "", fmt.Errorf("many2many field %q on %s missing relation table metadata", fieldDef.Name, parentModel)
	}
	if err := ValidateFieldName(relTable); err != nil {
		return "", err
	}
	if err := ValidateFieldName(leftCol); err != nil {
		return "", err
	}
	parentTbl, err := QuotedTableName(parentModel)
	if err != nil {
		return "", err
	}
	relTbl := quoteIdent(relTable)
	leftQuoted := quoteIdent(leftCol)
	parentIDCol := parentTbl + "." + quoteIdent("id")
	exists := fmt.Sprintf("EXISTS (SELECT 1 FROM %s rel WHERE rel.%s = %s)", relTbl, leftQuoted, parentIDCol)
	if wantSet {
		return exists, nil
	}
	return "NOT " + exists, nil
}

func cellIsFalsyEmpty(modelName, fieldName string, cellValue interface{}) bool {
	if fieldName != "" && modelName != "" {
		if fieldDef, kind, ok := lookupFalsyField(modelName, fieldName); ok {
			switch kind {
			case falsyNull:
				if cellValue == nil {
					return true
				}
				if id, ok := CoerceInt64(cellValue); ok {
					return id == 0
				}
				return false
			case falsyEmpty:
				return cellValue == nil || strings.TrimSpace(AsString(cellValue)) == ""
			case falsyNullOrZero:
				if cellValue == nil {
					return true
				}
				if n, ok := CoerceInt64(cellValue); ok {
					return n == 0
				}
				f, ok := toFloat64(cellValue)
				return ok && f == 0
			case falsyRelOne2Many, falsyRelMany2Many:
				return relationalCellEmpty(cellValue)
			case falsyNone:
				if fieldDef.Type == Boolean {
					return !AsBool(cellValue)
				}
			}
		}
	}
	return genericCellEmpty(cellValue)
}

func relationalCellEmpty(cellValue interface{}) bool {
	switch v := cellValue.(type) {
	case nil:
		return true
	case []interface{}:
		return len(v) == 0
	case []int:
		return len(v) == 0
	case []int64:
		return len(v) == 0
	case []map[string]interface{}:
		return len(v) == 0
	default:
		if id, ok := CoerceInt64(cellValue); ok {
			return id == 0
		}
		return strings.TrimSpace(AsString(cellValue)) == ""
	}
}

func genericCellEmpty(cellValue interface{}) bool {
	if cellValue == nil {
		return true
	}
	if relationalCellEmpty(cellValue) {
		return true
	}
	if s, ok := cellValue.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	if !AsBool(cellValue) {
		return true
	}
	return false
}

func cellMatchesFalsy(modelName, fieldName, op string, cellValue interface{}, wantBool bool) (matched bool, handled bool) {
	wantSet, ok := falsyWantSet(op, wantBool)
	if !ok {
		return false, false
	}
	if modelName != "" && fieldName != "" {
		if inst, has := Registry[modelName]; has && inst != nil {
			for _, fieldDef := range inst.Fields() {
				if fieldDef.Name != fieldName {
					continue
				}
				if fieldDef.Type == Boolean {
					got := AsBool(cellValue)
					if strings.TrimSpace(strings.ToLower(op)) == "!=" {
						return got != wantBool, true
					}
					return got == wantBool, true
				}
				kind := falsyKindForField(fieldDef)
				if kind == falsyNone {
					return false, false
				}
				empty := cellIsFalsyEmpty(modelName, fieldName, cellValue)
				if wantSet {
					return !empty, true
				}
				return empty, true
			}
		}
	}
	empty := genericCellEmpty(cellValue)
	if wantSet {
		return !empty, true
	}
	return empty, true
}
