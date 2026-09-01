package modelreg

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"sumeru/core/modelmeta"
	"sumeru/core/orm"
)

// MustRegister registers struct-based models for module. Called from generated code only.
func MustRegister(module string, models ...any) {
	module = strings.TrimSpace(module)
	if module == "" {
		panic("modelreg.MustRegister: module name is empty")
	}

	ctx := &registerCtx{
		typeNames: make(map[string]string),
		byName:    make(map[string]reflect.Type),
	}
	entries := make([]modelEntry, 0, len(models))
	for _, sample := range models {
		structType, ok := modelStructType(sample)
		if !ok {
			panic(fmt.Sprintf("modelreg.MustRegister: %T is not a pointer to struct", sample))
		}
		spec, err := modelmeta.ModelSpecFromStruct(structType)
		if err != nil {
			panic(fmt.Sprintf("modelreg.MustRegister: %v", err))
		}
		if spec.Name == "" || spec.Name == "-" {
			continue
		}
		ctx.index(structType, spec)
		entries = append(entries, modelEntry{structType: structType, spec: spec})
	}

	for _, entry := range entries {
		fields, err := reflectFields(ctx, entry.structType)
		if err != nil {
			panic(fmt.Sprintf("modelreg: model %s: %v", entry.spec.Name, err))
		}
		rm := &reflectedModel{name: entry.spec.Name, fields: fields}
		if entry.spec.Extend {
			extendRegisteredModel(entry.spec.Name, rm.fields, module)
			continue
		}
		orm.RegisterModelWithModule(rm, module)
	}
}

type modelEntry struct {
	structType reflect.Type
	spec       modelmeta.ModelSpec
}

type registerCtx struct {
	pkgDir    string
	typeNames map[string]string
	byName    map[string]reflect.Type
}

func (ctx *registerCtx) index(structType reflect.Type, spec modelmeta.ModelSpec) {
	ctx.typeNames[structType.String()] = spec.Name
	ctx.byName[structType.Name()] = structType
	if ctx.pkgDir == "" {
		ctx.pkgDir = pkgDirForStructType(structType)
	}
}

type reflectedModel struct {
	name   string
	fields []orm.FieldDefinition
}

func (m *reflectedModel) ModelName() string { return m.name }

func (m *reflectedModel) Fields() []orm.FieldDefinition { return m.fields }

func modelStructType(sample any) (reflect.Type, bool) {
	if sample == nil {
		return nil, false
	}
	rt := reflect.TypeOf(sample)
	if rt.Kind() != reflect.Pointer || rt.Elem().Kind() != reflect.Struct {
		return nil, false
	}
	return rt.Elem(), true
}

func (m *reflectedModel) mergeFields(extra []orm.FieldDefinition) error {
	for _, f := range extra {
		for _, existing := range m.fields {
			if existing.Name == f.Name {
				return fmt.Errorf("duplicate field %q on model %s", f.Name, m.name)
			}
		}
		m.fields = append(m.fields, f)
	}
	return nil
}

func extendRegisteredModel(targetName string, extra []orm.FieldDefinition, module string) {
	existing := orm.RegistryModel(targetName)
	if existing == nil {
		panic(fmt.Sprintf("modelreg: inherit=%s: target model is not registered", targetName))
	}
	rm, ok := existing.(*reflectedModel)
	if !ok {
		panic(fmt.Sprintf("modelreg: inherit=%s: cannot extend non-reflected model %T", targetName, existing))
	}
	if err := rm.mergeFields(extra); err != nil {
		panic(fmt.Sprintf("modelreg: inherit=%s: %v", targetName, err))
	}
	orm.RecordModelExtendedBy(targetName, module)
}

func reflectFields(ctx *registerCtx, structType reflect.Type) ([]orm.FieldDefinition, error) {
	var out []orm.FieldDefinition
	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		if !f.IsExported() || modelmeta.IsEmbeddedModelMeta(f) {
			continue
		}
		fieldDef, err := fieldFromStructField(ctx, f)
		if err != nil {
			return nil, err
		}
		out = append(out, fieldDef)
	}
	return out, nil
}

func fieldFromStructField(ctx *registerCtx, f reflect.StructField) (orm.FieldDefinition, error) {
	tags, err := modelmeta.ParseFieldTag(string(f.Tag.Get("sumeru")))
	if err != nil {
		return orm.FieldDefinition{}, fmt.Errorf("field %s: %w", f.Name, err)
	}

	name := modelmeta.FieldNameFromGo(f.Name)
	if tags.Column != "" {
		name = tags.Column
	}

	label := tags.Label
	if label == "" {
		label = modelmeta.LabelFromGo(f.Name)
	}

	marker := markerBaseName(f.Type)
	fieldType, widget, err := mapMarkerType(marker)
	if err != nil {
		return orm.FieldDefinition{}, fmt.Errorf("field %s: %w", f.Name, err)
	}

	fieldDef := orm.FieldDefinition{
		Name:      name,
		Type:      fieldType,
		Required:  tags.Required,
		Unique:    tags.Unique,
		Index:     tags.Index,
		Readonly:  tags.Readonly,
		String:    label,
		Help:      tags.Help,
		Size:      tags.Size,
		Precision: tags.Precision,
		Scale:     tags.Scale,
		OnDelete:  tags.OnDelete,
		Widget:    widget,
		Min:       tags.Min,
		Max:       tags.Max,
		Currency:  tags.Currency,
		Domain:    tags.Domain,
		Groups:    tags.Groups,
		Related:   tags.Related,
		Compute:   tags.Compute,
	}

	if tags.Related != "" {
		fieldDef.RelatedStore = tags.Store
		if !tags.Store {
			fieldDef.Virtual = true
		}
	}
	if tags.Compute != "" {
		fieldDef.ComputeStore = tags.Store
		if tags.Store {
			fieldDef.Readonly = true
		} else {
			fieldDef.Virtual = true
		}
	}
	if tags.Column != "" && tags.Column != name {
		fieldDef.Column = tags.Column
	}
	if tags.Default != "" {
		fieldDef.DefaultVal = parseDefault(tags.Default, fieldType)
	}

	if tags.Selection != "" {
		fieldDef.Type = orm.Selection
		fieldDef.Selection = parseSelection(tags.Selection)
	} else if marker == "Selection" {
		fieldDef.Type = orm.Selection
		if opts := selectionForFieldType(f.Type); len(opts) > 0 {
			fieldDef.Selection = opts
		} else if typeName := typeShortName(selectionTypeKey(f.Type)); typeName != "" {
			fieldDef.Selection = selectionOptionsFromPackage(ctx.pkgDir, typeName)
		}
	}

	switch fieldDef.Type {
	case orm.Many2One, orm.One2Many, orm.Many2Many:
		comodel, err := ctx.resolveComodel(f.Type, tags)
		if err != nil {
			return orm.FieldDefinition{}, fmt.Errorf("field %s: %w", f.Name, err)
		}
		fieldDef.Relation = comodel
		if fieldDef.Type == orm.Many2Many {
			fieldDef.RelationTable = tags.Table
			fieldDef.Column1 = tags.Left
			fieldDef.Column2 = tags.Right
		}
	case orm.Integer:
		if marker == "Many2OneReference" && tags.ModelField != "" {
			fieldDef.RelationModelField = tags.ModelField
		}
	}

	if fieldDef.Type == orm.Numeric && marker == "Money" {
		if fieldDef.Precision <= 0 {
			fieldDef.Precision = 18
		}
		if fieldDef.Scale <= 0 {
			fieldDef.Scale = 2
		}
	}
	if fieldDef.Type == orm.Char && marker == "UUID" && fieldDef.Size <= 0 {
		fieldDef.Size = 36
	}

	return fieldDef, nil
}

type markerType struct {
	fieldType orm.FieldType
	widget    string
}

var markerTypeByName = map[string]markerType{
	"String":            {orm.Char, ""},
	"Text":              {orm.Text, ""},
	"HTML":              {orm.Text, "html"},
	"Email":             {orm.Char, "email"},
	"Phone":             {orm.Char, "phone"},
	"URL":               {orm.Char, "url"},
	"UUID":              {orm.Char, ""},
	"Boolean":           {orm.Boolean, ""},
	"Integer":           {orm.Integer, ""},
	"Float":             {orm.Float, "float"},
	"Float64":           {orm.Float64, "float"},
	"Numeric":           {orm.Numeric, "numeric"},
	"Money":             {orm.Numeric, "monetary"},
	"Date":              {orm.Date, "date"},
	"Time":              {orm.Text, "time"},
	"DateTime":          {orm.DateTime, "datetime"},
	"Duration":          {orm.Numeric, "duration"},
	"Json":              {orm.Json, "json"},
	"Binary":            {orm.Text, "binary"},
	"Image":             {orm.Text, "image"},
	"Reference":         {orm.Char, "reference"},
	"Many2OneReference": {orm.Integer, "many2one_reference"},
	"Many2One":          {orm.Many2One, ""},
	"One2Many":          {orm.One2Many, ""},
	"Many2Many":         {orm.Many2Many, ""},
	"Selection":         {orm.Selection, ""},
}

func mapMarkerType(marker string) (orm.FieldType, string, error) {
	if m, ok := markerTypeByName[marker]; ok {
		return m.fieldType, m.widget, nil
	}
	return "", "", fmt.Errorf("unsupported field type %s", marker)
}

func markerBaseName(t reflect.Type) string {
	s := t.String()
	if i := strings.Index(s, "["); i >= 0 {
		s = s[:i]
	}
	return typeShortName(s)
}

func genericArgString(t reflect.Type) string {
	s := t.String()
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start < 0 || end <= start {
		return ""
	}
	return strings.TrimSpace(s[start+1 : end])
}

func typeShortName(typeArg string) string {
	typeArg = strings.TrimSpace(typeArg)
	if typeArg == "" {
		return ""
	}
	if dot := strings.LastIndex(typeArg, "."); dot >= 0 {
		return typeArg[dot+1:]
	}
	return typeArg
}

func (ctx *registerCtx) resolveComodel(fieldType reflect.Type, tags modelmeta.FieldTags) (string, error) {
	if comodel := strings.TrimSpace(tags.Comodel); comodel != "" {
		return comodel, nil
	}

	arg := genericArgString(fieldType)
	if arg == "" {
		return "", fmt.Errorf("relation field requires type parameter or comodel tag")
	}
	short := typeShortName(arg)
	if short == "Any" {
		return "", fmt.Errorf("relation on %s requires comodel tag", markerBaseName(fieldType))
	}

	if name, ok := ctx.typeNames[arg]; ok {
		return name, nil
	}
	if rt, ok := ctx.byName[short]; ok {
		if name, ok := ctx.typeNames[rt.String()]; ok {
			return name, nil
		}
	}
	for typeStr, name := range ctx.typeNames {
		if typeShortName(typeStr) == short {
			return name, nil
		}
	}

	return modelmeta.ModelNameFromGo(short), nil
}

func parseSelection(raw string) [][]string {
	var out [][]string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, label, hasLabel := strings.Cut(part, ":")
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if !hasLabel || strings.TrimSpace(label) == "" {
			label = modelmeta.LabelFromGo(key)
		} else {
			label = strings.TrimSpace(label)
		}
		out = append(out, []string{key, label})
	}
	return out
}

func parseDefault(raw string, fieldType orm.FieldType) interface{} {
	raw = strings.TrimSpace(raw)
	switch fieldType {
	case orm.Boolean:
		switch strings.ToLower(raw) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		default:
			return raw == "true"
		}
	case orm.Integer, orm.Many2One:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return raw
		}
		return n
	case orm.Float, orm.Float64, orm.Numeric:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return raw
		}
		return f
	default:
		return raw
	}
}
