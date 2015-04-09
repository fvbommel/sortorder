package util

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func debugf(f string, vs ...interface{}) {
	//~ log.Printf(f, vs...)
}

// ShortRegexpString tries to construct a short regexp that matches exactly the
// provided strings and nothing else. It assumes the strings contain no special
// values.
func ShortRegexpString(vs ...string) (res string) {

	defer func(s string) {
		debugf("ShortRegexpString(%s) = %#q", s, res)
	}(fmt.Sprintf("%#q", vs))

	switch len(vs) {
	case 0:
		return "$.^" // Unmatchable?
	case 1:
		return regexp.QuoteMeta(vs[0])
	}

	recurse := func(prefix string, data commonSs) (result string) {
		debugf("recurse(%#q, %v) on %#q", prefix, data, vs)
		if data.start > 0 {
			result += ShortRegexpString(vs[:data.start]...) + "|"
		}
		debugf("%v/%#q: %v\n", vs, prefix, data)
		suffixes := make([]string, data.end-data.start)
		for i := data.start; i < data.end; i++ {
			suffixes[i-data.start] = vs[i][len(prefix):]
		}
		middle := ShortRegexpString(suffixes...)
		debugf(">> ShortRegexpString(%#q) = %#q", suffixes, middle)
		result += fmt.Sprintf("%s(%s)", prefix, middle)
		if data.end < len(vs) {
			result += "|" + ShortRegexpString(vs[data.end:]...)
		}
		return result
	}

	bestCost := 1
	for _, v := range vs {
		bestCost += len(v) + 1
	}

	found := false
	prefixes := commonPrefixes(vs, 1)
	best := ""
	for k, v := range prefixes {
		str := recurse(k, v)
		if len(str) < bestCost {
			bestCost = len(str)
			best = str
			found = true
		}
	}
	if found {
		return best
	}

	// Last resort
	quoted := make([]string, len(vs))
	for i := range vs {
		quoted[i] = regexp.QuoteMeta(vs[i])
	}
	return strings.Join(quoted, "|")
}

type commonSs struct {
	start, end int
}

// commonPrefixes returns a map from prefixes to number of occurrences.
func commonPrefixes(vs []string, minLength int) (result map[string]commonSs) {
	sort.Strings(vs)
	result = make(map[string]commonSs)
	for i := 0; i < len(vs)-1; i++ {
		j := i + 1
		k := 0
		for ; k < len(vs[i]) && k < len(vs[j]); k++ {
			if vs[i][k] != vs[j][k] {
				break
			}
		}
		if k < minLength {
			continue
		}
		prefix := vs[i][:k]
		if _, exists := result[prefix]; !exists {
			first := prefixStart(vs[:i], prefix)
			debugf("prefixStart(%#q, %#q) == %v", vs[:i], prefix, first)
			// prefixEnd(vs, prefix) - first + 1
			// == prefixEnd(vs[first:], prefix) + 1
			// == prefixEnd(vs[first+1:], prefix) + 2
			end := first + 1 + prefixEnd(vs[first+1:], prefix)
			result[prefix] = commonSs{
				first, end,
			}
			debugf("prefixEnd(%#q, %#q) == %v", vs, prefix, result[prefix].end)
		}
	}
	return result
}

func prefixStart(vs []string, prefix string) int {
	if prefix == "" {
		return 0
	}
	return findFirst(vs, func(s string) bool {
		return strings.HasPrefix(s, prefix)
	})
}

func prefixEnd(vs []string, prefix string) int {
	if prefix == "" {
		return len(vs)
	}
	debugf("prefixEnd(%v, %#q)", vs, prefix)
	return findFirst(vs, func(s string) bool {
		return !strings.HasPrefix(s, prefix)
	})
}

func findFirst(vs []string, predicate func(string) bool) int {
	l, h := -1, len(vs)
	// Invariant: vs[l] does not match, vs[h] does.
	// -1 and len(vs) are sentinal values, never tested but assumed to mismatch and match, respectively.
	for l+1 < h {
		m := (l + h) / 2 // Must now be a valid value
		debugf("%d %d %d", l, m, h)
		if predicate(vs[m]) {
			h = m
		} else {
			l = m
		}
	}
	return h
}
