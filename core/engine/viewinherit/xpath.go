package viewinherit

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type xpathOp struct {
	Expr     string
	Position string
	Inner    string
}

var xpathBlockRe = regexp.MustCompile(`(?s)<xpath\s+expr="([^"]+)"\s+position="([^"]+)"\s*>(.*?)</xpath>`)
var xpathBlockReSingle = regexp.MustCompile(`(?s)<xpath\s+expr='([^']+)'\s+position='([^']+)'\s*>(.*?)</xpath>`)
var fieldNameFromExpr = regexp.MustCompile(`@name=['"]([^'"]+)['"]`)
var xpathTargetRe = regexp.MustCompile(`//(field|button|group|sheet|header|notebook|page|div|label|separator|filter|search|tree|list|form|kanban)\[@([a-zA-Z_:][\w-]*)=['"]([^'"]+)['"]\]`)
var xpathTagOnlyRe = regexp.MustCompile(`//(field|button|group|sheet|header|notebook|page|div|label|separator|filter|search|tree|list|form|kanban)\s*$`)
var dataWrapperRe = regexp.MustCompile(`(?s)^\s*<data[^>]*>(.*)</data>\s*$`)
var attributeOpRe = regexp.MustCompile(`(?s)<attribute\s+name=['"]([^'"]+)['"]\s*>(.*?)</attribute>`)

var regexCache sync.Map

// ApplyInheritArch parses <xpath> blocks in inheritFragment and applies them to parentArch.
// Supported targets: //field|button|group|sheet|header|notebook|page|div|…[@attr='value']
// Positions: after|before|inside|replace|attributes
func ApplyInheritArch(parentArch, inheritFragment string) (string, error) {
	arch := parentArch
	frag := stripDataWrapper(inheritFragment)
	ops := parseXPaths(frag)
	if len(ops) == 0 && strings.TrimSpace(frag) != "" {
		return arch, fmt.Errorf("no <xpath> blocks found in inherit arch")
	}
	for _, op := range ops {
		var err error
		arch, err = applyOne(arch, op)
		if err != nil {
			return arch, err
		}
	}
	return arch, nil
}

func parseXPaths(s string) []xpathOp {
	var out []xpathOp
	for _, re := range []*regexp.Regexp{xpathBlockRe, xpathBlockReSingle} {
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			out = append(out, xpathOp{Expr: m[1], Position: m[2], Inner: m[3]})
		}
	}
	return out
}

func xpathTargetFromExpr(expr string) (tag, attrName, attrVal string, err error) {
	expr = strings.TrimSpace(expr)
	if m := xpathTargetRe.FindStringSubmatch(expr); len(m) >= 4 {
		return strings.ToLower(m[1]), m[2], m[3], nil
	}
	if m := xpathTagOnlyRe.FindStringSubmatch(expr); len(m) >= 2 {
		return strings.ToLower(m[1]), "", "", nil
	}
	m := fieldNameFromExpr.FindStringSubmatch(expr)
	if len(m) >= 2 {
		return "field", "name", m[1], nil
	}
	return "", "", "", fmt.Errorf("unsupported xpath expr (use //tag[@attr='…'] or //field[@name='…']): %q", expr)
}

func cachedRegex(key string, build func() *regexp.Regexp) *regexp.Regexp {
	if v, ok := regexCache.Load(key); ok {
		return v.(*regexp.Regexp)
	}
	re := build()
	actual, _ := regexCache.LoadOrStore(key, re)
	return actual.(*regexp.Regexp)
}

func openingTagRe(tag, attrName, attrVal string) *regexp.Regexp {
	key := "open|" + tag + "|" + attrName + "|" + attrVal
	return cachedRegex(key, func() *regexp.Regexp {
		if attrName == "" {
			return regexp.MustCompile(`<` + tag + `(?:\s[^>]*)?>`)
		}
		q := regexp.QuoteMeta(attrVal)
		a := regexp.QuoteMeta(attrName)
		return regexp.MustCompile(`<` + tag + `\s+[^>]*\b` + a + `=(?:"` + q + `"|'` + q + `')[^>]*>`)
	})
}

func attrValuePattern(attrName, attrVal string) string {
	q := regexp.QuoteMeta(attrVal)
	a := regexp.QuoteMeta(attrName)
	return `\b` + a + `=(?:"` + q + `"|'` + q + `')`
}

func findElementSpan(arch, tag, attrName, attrVal string) (start, end int, ok bool) {
	openRe := openingTagRe(tag, attrName, attrVal)
	loc := openRe.FindStringIndex(arch)
	if loc == nil {
		return 0, 0, false
	}
	openTag := arch[loc[0]:loc[1]]
	if strings.HasSuffix(strings.TrimSpace(openTag), "/>") {
		return loc[0], loc[1], true
	}
	closeAt, found := findMatchingCloseTag(arch, loc[1], tag)
	if !found {
		return 0, 0, false
	}
	closeEnd := closeAt + len("</"+tag+">")
	return loc[0], closeEnd, true
}

func findMatchingCloseTag(arch string, openEnd int, tag string) (closeStart int, ok bool) {
	openNeedle := "<" + tag
	closeNeedle := "</" + tag + ">"
	pos := openEnd
	depth := 1
	for pos < len(arch) && depth > 0 {
		nextOpen := strings.Index(arch[pos:], openNeedle)
		nextClose := strings.Index(arch[pos:], closeNeedle)
		if nextClose < 0 {
			return 0, false
		}
		if nextOpen >= 0 && nextOpen < nextClose {
			absOpen := pos + nextOpen
			if isXMLTagOpen(arch, absOpen, tag) {
				depth++
				pos = absOpen + len(openNeedle)
				continue
			}
		}
		absClose := pos + nextClose
		depth--
		if depth == 0 {
			return absClose, true
		}
		pos = absClose + len(closeNeedle)
	}
	return 0, false
}

func isXMLTagOpen(arch string, idx int, tag string) bool {
	prefix := "<" + tag
	if idx+len(prefix) > len(arch) || arch[idx:idx+len(prefix)] != prefix {
		return false
	}
	if idx+len(prefix) >= len(arch) {
		return false
	}
	switch arch[idx+len(prefix)] {
	case ' ', '>', '/':
		return true
	default:
		return false
	}
}

func applyOne(arch string, op xpathOp) (string, error) {
	tag, attrName, attrVal, err := xpathTargetFromExpr(strings.TrimSpace(op.Expr))
	if err != nil {
		return arch, err
	}
	pos := strings.ToLower(strings.TrimSpace(op.Position))
	inner := strings.TrimSpace(op.Inner)

	switch pos {
	case "after":
		spanStart, spanEnd, ok := findElementSpan(arch, tag, attrName, attrVal)
		if !ok {
			return arch, fmt.Errorf("inherit xpath: %s not found for position=after", op.Expr)
		}
		_ = spanStart
		insert := inner
		if tag == "field" && attrName == "name" && !strings.HasPrefix(strings.TrimSpace(insert), "<") {
			insert = "<field name=\"" + insert + "\"/>"
		}
		return arch[:spanEnd] + insert + arch[spanEnd:], nil
	case "before":
		spanStart, _, ok := findElementSpan(arch, tag, attrName, attrVal)
		if !ok {
			return arch, fmt.Errorf("inherit xpath: %s not found for position=before", op.Expr)
		}
		return arch[:spanStart] + inner + arch[spanStart:], nil
	case "replace":
		spanStart, spanEnd, ok := findElementSpan(arch, tag, attrName, attrVal)
		if !ok {
			return arch, fmt.Errorf("inherit xpath: %s not found for position=replace", op.Expr)
		}
		return arch[:spanStart] + inner + arch[spanEnd:], nil
	case "inside":
		if attrName != "" {
			openRe := openingTagRe(tag, attrName, attrVal)
			loc := openRe.FindStringIndex(arch)
			if loc == nil {
				return arch, fmt.Errorf("inherit xpath: %s not found for position=inside", op.Expr)
			}
			insertAt := loc[1]
			closeAt, ok := findMatchingCloseTag(arch, insertAt, tag)
			if !ok {
				return arch, fmt.Errorf("inherit xpath: no </%s> for position=inside on %s", tag, op.Expr)
			}
			return arch[:closeAt] + inner + arch[closeAt:], nil
		}
		i := strings.LastIndex(arch, "</view>")
		if i < 0 {
			return arch, fmt.Errorf("inherit xpath: no </view> in arch for position=inside")
		}
		return arch[:i] + inner + arch[i:], nil
	case "attributes":
		return applyAttributes(arch, tag, attrName, attrVal, inner)
	default:
		return arch, fmt.Errorf("unsupported xpath position %q", op.Position)
	}
}

func applyAttributes(arch, tag, matchAttr, matchVal, inner string) (string, error) {
	re := openingTagRe(tag, matchAttr, matchVal)
	loc := re.FindStringIndex(arch)
	if loc == nil {
		return arch, fmt.Errorf("inherit xpath: %q not found for position=attributes", tag)
	}
	old := arch[loc[0]:loc[1]]
	attrs := parseAttributeOps(inner)
	if len(attrs) == 0 {
		return arch, fmt.Errorf("inherit xpath: no <attribute> elements for position=attributes")
	}
	next := old
	for name, val := range attrs {
		next = upsertXMLAttr(next, name, val)
	}
	return arch[:loc[0]] + next + arch[loc[1]:], nil
}

func parseAttributeOps(inner string) map[string]string {
	out := map[string]string{}
	for _, m := range attributeOpRe.FindAllStringSubmatch(inner, -1) {
		name := strings.TrimSpace(m[1])
		if name == "" {
			continue
		}
		out[name] = strings.TrimSpace(m[2])
	}
	return out
}

func upsertXMLAttr(openTag, name, value string) string {
	nameRe := regexp.MustCompile(`\s` + regexp.QuoteMeta(name) + `=(?:"[^"]*"|'[^']*')`)
	replacement := ` ` + name + `="` + value + `"`
	if nameRe.MatchString(openTag) {
		return nameRe.ReplaceAllString(openTag, replacement)
	}
	if strings.HasSuffix(openTag, "/>") {
		return strings.TrimSuffix(openTag, "/>") + replacement + `/>`
	}
	if strings.HasSuffix(openTag, ">") {
		return strings.TrimSuffix(openTag, ">") + replacement + `>`
	}
	return openTag + replacement
}

func stripDataWrapper(s string) string {
	s = strings.TrimSpace(s)
	if m := dataWrapperRe.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return s
}
