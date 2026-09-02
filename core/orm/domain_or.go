package orm

import "fmt"

// splitDomainORPrefix returns orCount and leaf triples from a prefix-Polish OR domain.
// ok is false when OR markers are present but the leaf count does not match.
func splitDomainORPrefix(domain [][]interface{}) (orCount int, leaves [][]interface{}, ok bool) {
	for _, d := range domain {
		if len(d) == 1 && fmt.Sprint(d[0]) == "|" {
			orCount++
			continue
		}
		break
	}
	if orCount == 0 {
		return 0, nil, true
	}
	leaves = domain[orCount:]
	if len(leaves) != orCount+1 {
		return orCount, leaves, false
	}
	return orCount, leaves, true
}
