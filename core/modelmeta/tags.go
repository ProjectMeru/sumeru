package modelmeta

import (
	"fmt"
	"strconv"
	"strings"
)

// FieldTags holds parsed sumeru struct tag options for one field.
type FieldTags struct {
	Model        string
	Inherit      string
	Required     bool
	Unique       bool
	Index        bool
	Readonly     bool
	Store        bool
	Default      string
	Column       string
	Size         int
	Precision    int
	Scale        int
	Label        string
	Comodel      string
	OnDelete     string
	Inverse      string
	Table        string
	Left         string
	Right        string
	Selection    string
	ModelField   string
	Min          *float64
	Max          *float64
	Help         string
	Currency     string
	Domain       string
	Groups       string
	Related      string
	Compute      string
}

// ParseModelTag parses the sumeru tag on an embedded ModelMeta.
// For inherit= tags, returns the inherited model name.
func ParseModelTag(tag string) (modelName string, err error) {
	tags, err := parseSumeruTag(tag)
	if err != nil {
		return "", err
	}
	if tags.Model != "" && tags.Inherit != "" {
		return "", fmt.Errorf("model= and inherit= are mutually exclusive")
	}
	if tags.Inherit != "" {
		return tags.Inherit, nil
	}
	if tags.Model == "" {
		return "", nil
	}
	if tags.Model == "-" {
		return "-", nil
	}
	return tags.Model, nil
}

// ParseFieldTag parses a field's sumeru struct tag.
func ParseFieldTag(tag string) (FieldTags, error) {
	return parseSumeruTag(tag)
}

func parseSumeruTag(tag string) (FieldTags, error) {
	var out FieldTags
	tag = strings.TrimSpace(tag)
	if tag == "" || tag == "-" {
		return out, nil
	}
	if i := strings.Index(tag, "selection="); i >= 0 {
		out.Selection = strings.TrimSpace(tag[i+len("selection="):])
		tag = strings.TrimSpace(tag[:i])
		tag = strings.TrimSuffix(tag, ",")
	}
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, hasVal := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		if hasVal {
			val = strings.TrimSpace(val)
		}
		switch key {
		case "model":
			out.Model = val
		case "inherit":
			out.Inherit = val
		case "required":
			out.Required = true
		case "unique":
			out.Unique = true
		case "index":
			out.Index = true
		case "readonly":
			out.Readonly = true
		case "store":
			out.Store = true
		case "default":
			out.Default = val
		case "column":
			out.Column = val
		case "size":
			n, err := strconv.Atoi(val)
			if err != nil {
				return out, fmt.Errorf("invalid size=%q", val)
			}
			out.Size = n
		case "precision":
			n, err := strconv.Atoi(val)
			if err != nil {
				return out, fmt.Errorf("invalid precision=%q", val)
			}
			out.Precision = n
		case "scale":
			n, err := strconv.Atoi(val)
			if err != nil {
				return out, fmt.Errorf("invalid scale=%q", val)
			}
			out.Scale = n
		case "string", "label":
			out.Label = val
		case "comodel", "relation":
			out.Comodel = val
		case "ondelete":
			out.OnDelete = val
		case "inverse":
			out.Inverse = val
		case "table":
			out.Table = val
		case "left":
			out.Left = val
		case "right":
			out.Right = val
		case "model_field":
			out.ModelField = val
		case "min":
			n, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return out, fmt.Errorf("invalid min=%q", val)
			}
			out.Min = &n
		case "max":
			n, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return out, fmt.Errorf("invalid max=%q", val)
			}
			out.Max = &n
		case "help":
			out.Help = val
		case "currency":
			out.Currency = val
		case "domain":
			out.Domain = val
		case "groups":
			out.Groups = val
		case "related":
			out.Related = val
		case "compute":
			out.Compute = val
		default:
			return out, fmt.Errorf("unknown sumeru tag %q", key)
		}
	}
	return out, nil
}
