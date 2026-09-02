package modelmeta

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	selectionTagPrefix       = "selection="
	defaultInSelectionSuffix = ",default="
)

// FieldTags holds parsed sumeru struct tag options for one field.
type FieldTags struct {
	Model      string
	Inherit    string
	Inherits   string // delegation parent model (Enterprise _inherits)
	Required   bool
	Unique     bool
	Index      bool
	Readonly   bool
	Store      bool
	Default    string
	Column     string
	Size       int
	Precision  int
	Scale      int
	Label      string
	Comodel    string
	OnDelete   string
	Inverse    string
	Table      string
	Left       string
	Right      string
	Selection  string
	ModelField string
	Min        *float64
	Max        *float64
	Help       string
	Currency   string
	Domain     string
	Groups     string
	Related    string
	Compute    string
}

// ParseModelTag parses the sumeru tag on an embedded ModelMeta.
// For inherit= tags, returns the inherited model name.
func ParseModelTag(tag string) (modelName string, err error) {
	tags, err := parseSumeruTag(tag)
	if err != nil {
		return "", err
	}
	if err := validateModelInheritExclusive(tags); err != nil {
		return "", err
	}
	if tags.Inherit != "" {
		return tags.Inherit, nil
	}
	return tags.Model, nil
}

// ParseFieldTag parses a field's sumeru struct tag.
// Put selection= last when it contains commas; a trailing ,default=value on the
// selection payload is also accepted for backward compatibility.
func ParseFieldTag(tag string) (FieldTags, error) {
	return parseSumeruTag(tag)
}

func validateModelInheritExclusive(tags FieldTags) error {
	if tags.Model != "" && tags.Inherit != "" {
		return fmt.Errorf("model= and inherit= are mutually exclusive")
	}
	return nil
}

func parseSumeruTag(raw string) (FieldTags, error) {
	var tags FieldTags
	body, selection := peelSelectionTag(strings.TrimSpace(raw))
	selection, tailDefault := extractDefaultFromSelectionTail(selection)
	tags.Selection = selection

	seen := map[string]struct{}{}
	if tailDefault != "" {
		if err := applyTagOption(&tags, seen, "default", tailDefault); err != nil {
			return tags, err
		}
	}
	if body == "" || body == "-" {
		return tags, nil
	}

	for part := range strings.SplitSeq(body, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, hasValue := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		if hasValue {
			value = strings.TrimSpace(value)
		}
		if err := applyTagOption(&tags, seen, key, value); err != nil {
			return tags, err
		}
	}
	return tags, nil
}

// peelSelectionTag removes selection=... from the tag string.
// The selection value is kept intact (it may contain commas).
func peelSelectionTag(tag string) (body, selection string) {
	if tag == "" || tag == "-" {
		return tag, ""
	}
	index := strings.Index(tag, selectionTagPrefix)
	if index < 0 {
		return tag, ""
	}
	selection = strings.TrimSpace(tag[index+len(selectionTagPrefix):])
	body = strings.TrimSpace(tag[:index])
	body = strings.TrimSuffix(body, ",")
	return body, selection
}

// extractDefaultFromSelectionTail pulls a trailing ,default=value off a selection payload.
// Legacy tags sometimes place default after selection options in the same selection= value.
func extractDefaultFromSelectionTail(selection string) (trimmedSelection, defaultValue string) {
	index := strings.LastIndex(selection, defaultInSelectionSuffix)
	if index < 0 {
		return selection, ""
	}
	defaultValue = strings.TrimSpace(selection[index+len(defaultInSelectionSuffix):])
	if defaultValue == "" || strings.Contains(defaultValue, ",") {
		return selection, ""
	}
	return strings.TrimSpace(selection[:index]), defaultValue
}

func applyTagOption(tags *FieldTags, seen map[string]struct{}, key, value string) error {
	canonicalKey := canonicalTagKey(key)
	if _, ok := seen[canonicalKey]; ok {
		return fmt.Errorf("duplicate sumeru tag %q", key)
	}
	seen[canonicalKey] = struct{}{}
	return setTagOption(tags, key, value)
}

func canonicalTagKey(key string) string {
	switch key {
	case "string", "label":
		return "label"
	case "comodel", "relation":
		return "comodel"
	default:
		return key
	}
}

func setTagOption(tags *FieldTags, key, value string) error {
	switch key {
	case "model":
		tags.Model = value
	case "inherit":
		tags.Inherit = value
	case "inherits":
		tags.Inherits = value
	case "required":
		tags.Required = true
	case "unique":
		tags.Unique = true
	case "index":
		tags.Index = true
	case "readonly":
		tags.Readonly = true
	case "store":
		tags.Store = true
	case "default":
		tags.Default = value
	case "column":
		tags.Column = value
	case "size":
		n, err := parseTagInt("size", value)
		if err != nil {
			return err
		}
		tags.Size = n
	case "precision":
		n, err := parseTagInt("precision", value)
		if err != nil {
			return err
		}
		tags.Precision = n
	case "scale":
		n, err := parseTagInt("scale", value)
		if err != nil {
			return err
		}
		tags.Scale = n
	case "string", "label":
		tags.Label = value
	case "comodel", "relation":
		tags.Comodel = value
	case "ondelete":
		tags.OnDelete = value
	case "inverse":
		tags.Inverse = value
	case "table":
		tags.Table = value
	case "left":
		tags.Left = value
	case "right":
		tags.Right = value
	case "model_field":
		tags.ModelField = value
	case "min":
		n, err := parseTagFloat("min", value)
		if err != nil {
			return err
		}
		tags.Min = &n
	case "max":
		n, err := parseTagFloat("max", value)
		if err != nil {
			return err
		}
		tags.Max = &n
	case "help":
		tags.Help = value
	case "currency":
		tags.Currency = value
	case "domain":
		tags.Domain = value
	case "groups":
		tags.Groups = value
	case "related":
		tags.Related = value
	case "compute":
		tags.Compute = value
	default:
		return fmt.Errorf("unknown sumeru tag %q", key)
	}
	return nil
}

func parseTagInt(name, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q", name, value)
	}
	return n, nil
}

func parseTagFloat(name, value string) (float64, error) {
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q", name, value)
	}
	return n, nil
}
