package modelreg

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

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
		module:    module,
		typeNames: make(map[string]string),
		byName:    make(map[string]reflect.Type),
		specs:     make(map[string]modelmeta.ModelSpec),
	}

	for _, sample := range models {
		rt, ok := modelStructType(sample)
		if !ok {
			panic(fmt.Sprintf("modelreg.MustRegister: %T is not a pointer to struct", sample))
		}
		spec, err := modelSpecFromStruct(rt)
		if err != nil {
			panic(fmt.Sprintf("modelreg.MustRegister: %v", err))
		}
		name := spec.Name
		if name == "" || name == "-" {
			continue
		}
		ctx.typeNames[rt.String()] = name
		ctx.byName[rt.Name()] = rt
		ctx.specs[rt.String()] = spec
		if ctx.pkgDir == "" {
			ctx.pkgDir = pkgDirForStructType(rt)
		}
	}

	for _, sample := range models {
		rt, ok := modelStructType(sample)
		if !ok {
			continue
		}
		spec := ctx.specs[rt.String()]
		name := spec.Name
		if name == "" || name == "-" {
			continue
		}
		rm := buildReflectedModel(ctx, rt, name)
		if spec.Extend {
			extendRegisteredModel(name, rm.fields, module)
			continue
		}
		orm.RegisterModelWithModule(rm, module)
	}
}

type registerCtx struct {
	module    string
	pkgDir    string
	typeNames map[string]string
	byName    map[string]reflect.Type
	specs     map[string]modelmeta.ModelSpec
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

func modelNameFromStruct(st reflect.Type) (string, error) {
	return modelmeta.ModelNameFromStruct(st)
}

func modelSpecFromStruct(st reflect.Type) (modelmeta.ModelSpec, error) {
	return modelmeta.ModelSpecFromStruct(st)
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

func buildReflectedModel(ctx *registerCtx, st reflect.Type, modelName string) *reflectedModel {
	fields, err := reflectFields(ctx, st)
	if err != nil {
		panic(fmt.Sprintf("modelreg: model %s: %v", modelName, err))
	}
	return &reflectedModel{name: modelName, fields: fields}
}

func reflectFields(ctx *registerCtx, st reflect.Type) ([]orm.FieldDefinition, error) {
	var out []orm.FieldDefinition
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		if !f.IsExported() {
			continue
		}
		if isModelMetaField(f) {
			continue
		}
		fd, err := fieldFromStructField(ctx, f)
		if err != nil {
			return nil, err
		}
		out = append(out, fd)
	}
	return out, nil
}

func isModelMetaField(f reflect.StructField) bool {
	if !f.Anonymous {
		return false
	}
	t := f.Type
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t == reflect.TypeOf(modelmeta.ModelMeta{})
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

	ft, widget, err := mapMarkerType(f.Type)
	if err != nil {
		return orm.FieldDefinition{}, fmt.Errorf("field %s: %w", f.Name, err)
	}

	fd := orm.FieldDefinition{
		Name:      name,
		Type:      ft,
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
		fd.RelatedStore = tags.Store
		if !tags.Store {
			fd.Virtual = true
		}
	}
	if tags.Compute != "" {
		fd.ComputeStore = tags.Store
		if tags.Store {
			fd.Readonly = true
		} else {
			fd.Virtual = true
		}
	}

	if tags.Column != "" && tags.Column != name {
		fd.Column = tags.Column
	}

	if tags.Default != "" {
		fd.DefaultVal = parseDefault(tags.Default, ft)
	}

	if tags.Selection != "" {
		fd.Type = orm.Selection
		fd.Selection = parseSelection(tags.Selection)
	} else if markerBaseName(f.Type) == "Selection" {
		fd.Type = orm.Selection
		if opts := selectionForFieldType(f.Type); len(opts) > 0 {
			fd.Selection = opts
		} else if typeName := typeShortName(selectionTypeKey(f.Type)); typeName != "" {
			fd.Selection = selectionOptionsFromPackage(ctx.pkgDir, typeName)
		}
	}

	switch ft {
	case orm.Many2One:
		comodel, err := ctx.resolveComodel(f.Type, tags)
		if err != nil {
			return orm.FieldDefinition{}, fmt.Errorf("field %s: %w", f.Name, err)
		}
		fd.Relation = comodel
	case orm.One2Many:
		comodel, err := ctx.resolveComodel(f.Type, tags)
		if err != nil {
			return orm.FieldDefinition{}, fmt.Errorf("field %s: %w", f.Name, err)
		}
		fd.Relation = comodel
	case orm.Many2Many:
		comodel, err := ctx.resolveComodel(f.Type, tags)
		if err != nil {
			return orm.FieldDefinition{}, fmt.Errorf("field %s: %w", f.Name, err)
		}
		fd.Relation = comodel
		fd.RelationTable = tags.Table
		fd.Column1 = tags.Left
		fd.Column2 = tags.Right
	}

	if markerBaseName(f.Type) == "Many2OneReference" && tags.ModelField != "" {
		fd.RelationModelField = tags.ModelField
	}

	if fd.Type == orm.Numeric {
		if fd.Precision <= 0 && markerBaseName(f.Type) == "Money" {
			fd.Precision = 18
		}
		if fd.Scale <= 0 && markerBaseName(f.Type) == "Money" {
			fd.Scale = 2
		}
	}

	if fd.Type == orm.Char && markerBaseName(f.Type) == "UUID" && fd.Size <= 0 {
		fd.Size = 36
	}

	return fd, nil
}

func mapMarkerType(t reflect.Type) (orm.FieldType, string, error) {
	base := markerBaseName(t)
	switch base {
	case "String":
		return orm.Char, "", nil
	case "Text":
		return orm.Text, "", nil
	case "HTML":
		return orm.Text, "html", nil
	case "Email":
		return orm.Char, "email", nil
	case "Phone":
		return orm.Char, "phone", nil
	case "URL":
		return orm.Char, "url", nil
	case "UUID":
		return orm.Char, "", nil
	case "Boolean":
		return orm.Boolean, "", nil
	case "Integer":
		return orm.Integer, "", nil
	case "Float":
		return orm.Float, "float", nil
	case "Float64":
		return orm.Float64, "float", nil
	case "Numeric":
		return orm.Numeric, "numeric", nil
	case "Money":
		return orm.Numeric, "monetary", nil
	case "Date":
		return orm.Date, "date", nil
	case "Time":
		return orm.Text, "time", nil
	case "DateTime":
		return orm.DateTime, "datetime", nil
	case "Duration":
		return orm.Numeric, "duration", nil
	case "Json":
		return orm.Json, "json", nil
	case "Binary":
		return orm.Text, "binary", nil
	case "Image":
		return orm.Text, "image", nil
	case "Reference":
		return orm.Char, "reference", nil
	case "Many2OneReference":
		return orm.Integer, "many2one_reference", nil
	case "Many2One":
		return orm.Many2One, "", nil
	case "One2Many":
		return orm.One2Many, "", nil
	case "Many2Many":
		return orm.Many2Many, "", nil
	case "Selection":
		return orm.Selection, "", nil
	default:
		return "", "", fmt.Errorf("unsupported field type %s", t.String())
	}
}

func markerBaseName(t reflect.Type) string {
	s := t.String()
	if i := strings.Index(s, "["); i >= 0 {
		s = s[:i]
	}
	if dot := strings.LastIndex(s, "."); dot >= 0 {
		return s[dot+1:]
	}
	return s
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

	candidate := modelmeta.ModelNameFromGo(short)
	if orm.RegistryModel(candidate) != nil {
		return candidate, nil
	}
	return candidate, nil
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
			label = titleWords(key)
		} else {
			label = strings.TrimSpace(label)
		}
		out = append(out, []string{key, label})
	}
	return out
}

func titleWords(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	parts := strings.Fields(s)
	for i, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

func parseDefault(raw string, ft orm.FieldType) interface{} {
	raw = strings.TrimSpace(raw)
	switch ft {
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
	case orm.Float, orm.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return raw
		}
		return f
	case orm.Numeric:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return raw
		}
		return f
	default:
		return raw
	}
}
