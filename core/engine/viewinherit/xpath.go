package viewinherit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type xpathOp struct {
	Expr     string
	Position string
	Inner    string
}

type xpathTarget struct {
	Tag       string
	AttrName  string
	AttrVal   string
	ClassName string // set when expr uses hasclass('…')
	Index     int    // 0 = no [n] in expr; >0 = 1-based match index
}

type elementSpan struct {
	start   int
	end     int
	openEnd int
}

var xpathBlockRe = regexp.MustCompile(`(?s)<xpath\s+expr="([^"]+)"\s+position="([^"]+)"\s*>(.*?)</xpath>`)
var xpathBlockReSingle = regexp.MustCompile(`(?s)<xpath\s+expr='([^']+)'\s+position='([^']+)'\s*>(.*?)</xpath>`)
var fieldNameFromExpr = regexp.MustCompile(`@name=['"]([^'"]+)['"]`)
var xpathTagNames = `field|button|group|sheet|header|notebook|page|div|label|separator|filter|search|tree|list|form|kanban|section|document|report|title|table|row|p`

var xpathTargetRe = regexp.MustCompile(`//(` + xpathTagNames + `)\[@([a-zA-Z_:][\w-]*)=['"]([^'"]+)['"]\](?:\[(\d+)\])?`)
var xpathHasClassRe = regexp.MustCompile(`//(` + xpathTagNames + `)\[hasclass\(['"]([^'"]+)['"]\)\](?:\[(\d+)\])?`)
var xpathTagOnlyRe = regexp.MustCompile(`//(` + xpathTagNames + `)\s*$`)
var dataWrapperRe = regexp.MustCompile(`(?s)^\s*<data[^>]*>(.*)</data>\s*$`)
var attributeOpRe = regexp.MustCompile(`(?s)<attribute\s+name=['"]([^'"]+)['"]\s*>(.*?)</attribute>`)

// ApplyInheritArch parses <xpath> blocks in inheritFragment and applies them to parentArch.
// Limited XPath: //tag[@attr='val'][n], //tag[hasclass('cls')][n], //tag, and @name='field' shorthand.
// Positions: after|before|inside|replace|attributes. Multiple matches without [n] are an error.
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

func xpathOptionalIndex(m []string, idxPos int) (int, error) {
	if idxPos >= len(m) || strings.TrimSpace(m[idxPos]) == "" {
		return 0, nil
	}
	return parseXPathIndex(m[idxPos])
}

func parseXPathIndex(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("xpath index must be a positive integer, got %q", raw)
	}
	return n, nil
}

func parseXPathTarget(expr string) (xpathTarget, error) {
	expr = strings.TrimSpace(expr)
	if m := xpathHasClassRe.FindStringSubmatch(expr); len(m) >= 3 {
		idx, err := xpathOptionalIndex(m, 3)
		if err != nil {
			return xpathTarget{}, err
		}
		return xpathTarget{Tag: strings.ToLower(m[1]), ClassName: m[2], Index: idx}, nil
	}
	if m := xpathTargetRe.FindStringSubmatch(expr); len(m) >= 4 {
		idx, err := xpathOptionalIndex(m, 4)
		if err != nil {
			return xpathTarget{}, err
		}
		return xpathTarget{Tag: strings.ToLower(m[1]), AttrName: m[2], AttrVal: m[3], Index: idx}, nil
	}
	if m := xpathTagOnlyRe.FindStringSubmatch(expr); len(m) >= 2 {
		return xpathTarget{Tag: strings.ToLower(m[1]), Index: 0}, nil
	}
	m := fieldNameFromExpr.FindStringSubmatch(expr)
	if len(m) >= 2 {
		return xpathTarget{Tag: "field", AttrName: "name", AttrVal: m[1], Index: 0}, nil
	}
	return xpathTarget{}, fmt.Errorf("unsupported xpath expr (use //tag[@attr='…'] or //field[@name='…']): %q", expr)
}

func findAllElementSpans(arch string, target xpathTarget) []elementSpan {
	var spans []elementSpan
	tagNeedle := "<" + target.Tag
	pos := 0
	for pos < len(arch) {
		rel := strings.Index(arch[pos:], tagNeedle)
		if rel < 0 {
			break
		}
		abs := pos + rel
		if !isXMLTagOpen(arch, abs, target.Tag) {
			pos = abs + 1
			continue
		}
		openEnd, selfClosing, ok := findOpenTagEnd(arch, abs)
		if !ok {
			pos = abs + 1
			continue
		}
		openTag := arch[abs:openEnd]
		if !openTagMatchesTarget(openTag, target) {
			pos = openEnd
			continue
		}
		end := openEnd
		if !selfClosing {
			closeAt, found := findMatchingCloseTag(arch, openEnd, target.Tag)
			if !found {
				pos = openEnd
				continue
			}
			end = closeAt + len("</"+target.Tag+">")
		}
		spans = append(spans, elementSpan{start: abs, end: end, openEnd: openEnd})
		pos = end
	}
	return spans
}

func resolveElementSpan(arch string, target xpathTarget, expr string) (elementSpan, error) {
	spans := findAllElementSpans(arch, target)
	if len(spans) == 0 {
		return elementSpan{}, fmt.Errorf("inherit xpath: %s not found", expr)
	}
	if target.Index == 0 {
		if len(spans) > 1 {
			return elementSpan{}, fmt.Errorf("inherit xpath: ambiguous match (%d nodes) for %s; add a [n] index", len(spans), expr)
		}
		return spans[0], nil
	}
	if target.Index > len(spans) {
		return elementSpan{}, fmt.Errorf("inherit xpath: index [%d] out of range (%d matches) for %s", target.Index, len(spans), expr)
	}
	return spans[target.Index-1], nil
}

func findOpenTagEnd(arch string, start int) (openEnd int, selfClosing bool, ok bool) {
	if start >= len(arch) || arch[start] != '<' {
		return 0, false, false
	}
	var inQuote byte
	for i := start + 1; i < len(arch); i++ {
		c := arch[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
		case '>':
			inner := strings.TrimSpace(arch[start+1 : i])
			selfClosing = strings.HasSuffix(inner, "/")
			return i + 1, selfClosing, true
		}
	}
	return 0, false, false
}

func openTagMatchesTarget(openTag string, target xpathTarget) bool {
	attrs := parseXMLAttrs(openTag)
	if target.ClassName != "" {
		return classAttrHasToken(attrs["class"], target.ClassName)
	}
	if target.AttrName != "" {
		return attrs[target.AttrName] == target.AttrVal
	}
	return true
}

func parseXMLAttrs(openTag string) map[string]string {
	out := map[string]string{}
	rest := strings.TrimSpace(openTag)
	if !strings.HasPrefix(rest, "<") {
		return out
	}
	rest = strings.TrimSpace(rest[1:])
	i := 0
	for i < len(rest) && rest[i] != ' ' && rest[i] != '\t' && rest[i] != '\n' && rest[i] != '\r' && rest[i] != '>' && rest[i] != '/' {
		i++
	}
	rest = strings.TrimSpace(rest[i:])
	rest = strings.TrimSuffix(strings.TrimSpace(rest), ">")
	rest = strings.TrimSuffix(rest, "/")
	for len(rest) > 0 {
		eq := strings.IndexByte(rest, '=')
		if eq <= 0 {
			break
		}
		name := strings.TrimSpace(rest[:eq])
		valPart := strings.TrimSpace(rest[eq+1:])
		if len(valPart) == 0 {
			break
		}
		quote := valPart[0]
		if quote != '"' && quote != '\'' {
			break
		}
		endQuote := strings.IndexByte(valPart[1:], quote)
		if endQuote < 0 {
			break
		}
		out[name] = valPart[1 : 1+endQuote]
		rest = strings.TrimSpace(valPart[1+endQuote+1:])
	}
	return out
}

func classAttrHasToken(classAttr, token string) bool {
	if token == "" {
		return false
	}
	for _, part := range strings.Fields(classAttr) {
		if part == token {
			return true
		}
	}
	return false
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
	expr := strings.TrimSpace(op.Expr)
	target, err := parseXPathTarget(expr)
	if err != nil {
		return arch, err
	}
	pos := strings.ToLower(strings.TrimSpace(op.Position))
	inner := strings.TrimSpace(op.Inner)

	switch pos {
	case "after":
		span, err := resolveElementSpan(arch, target, expr)
		if err != nil {
			return arch, fmt.Errorf("%s for position=after", err.Error())
		}
		insert := inner
		if target.Tag == "field" && target.AttrName == "name" && !strings.HasPrefix(strings.TrimSpace(insert), "<") {
			insert = "<field name=\"" + insert + "\"/>"
		}
		return arch[:span.end] + insert + arch[span.end:], nil
	case "before":
		span, err := resolveElementSpan(arch, target, expr)
		if err != nil {
			return arch, fmt.Errorf("%s for position=before", err.Error())
		}
		return arch[:span.start] + inner + arch[span.start:], nil
	case "replace":
		span, err := resolveElementSpan(arch, target, expr)
		if err != nil {
			return arch, fmt.Errorf("%s for position=replace", err.Error())
		}
		return arch[:span.start] + inner + arch[span.end:], nil
	case "inside":
		span, err := resolveElementSpan(arch, target, expr)
		if err != nil {
			return arch, fmt.Errorf("%s for position=inside", err.Error())
		}
		if span.openEnd >= span.end {
			return arch, fmt.Errorf("inherit xpath: %s has no child slot for position=inside", expr)
		}
		closeAt, found := findMatchingCloseTag(arch, span.openEnd, target.Tag)
		if !found {
			return arch, fmt.Errorf("inherit xpath: no </%s> for position=inside on %s", target.Tag, expr)
		}
		return arch[:closeAt] + inner + arch[closeAt:], nil
	case "attributes":
		return applyAttributes(arch, target, inner, expr)
	default:
		return arch, fmt.Errorf("unsupported xpath position %q", op.Position)
	}
}

func applyAttributes(arch string, target xpathTarget, inner, expr string) (string, error) {
	span, err := resolveElementSpan(arch, target, expr)
	if err != nil {
		return arch, fmt.Errorf("%s for position=attributes", err.Error())
	}
	old := arch[span.start:span.openEnd]
	attrs := parseAttributeOps(inner)
	if len(attrs) == 0 {
		return arch, fmt.Errorf("inherit xpath: no <attribute> elements for position=attributes")
	}
	next := old
	for name, val := range attrs {
		next = upsertXMLAttr(next, name, val)
	}
	return arch[:span.start] + next + arch[span.openEnd:], nil
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
