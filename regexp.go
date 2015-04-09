package util

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var level = 0

func debugf(f string, vs ...interface{}) {
	// log.Printf(f, vs...)
}

// ShortRegexpString tries to construct a short regexp that matches exactly the
// provided strings and nothing else. It assumes the strings contain no special
// values.
func ShortRegexpString(vs ...string) (res string) {
	switch len(vs) {
	case 0:
		return "$.^" // Unmatchable?
	case 1:
		return regexp.QuoteMeta(vs[0])
	}

	level++
	defer func(s string) {
		level--
		debugf("%sShortRegexpString(%s) = %#q", strings.Repeat("\t", level), s, res)
	}(fmt.Sprintf("%#q", vs))

	recurse := func(substr string, data commonSs, isPrefix bool, trim func(string, string) string) (result string) {
		debugf("recurse(%#q, %v) on %#q", substr, data, vs)
		if data.start > 0 {
			result += ShortRegexpString(vs[:data.start]...) + "|"
		}
		debugf("%v/%#q: %v\n", vs, substr, data)
		suffixes := make([]string, data.end-data.start)
		for i := data.start; i < data.end; i++ {
			suffixes[i-data.start] = trim(vs[i], substr)
		}
		middle := ShortRegexpString(suffixes...)
		debugf(">> ShortRegexpString(%#q) = %#q", suffixes, middle)
		opt := ""
		if strings.HasPrefix(middle, "|") {
			middle = middle[1:]
			opt = "?"
		}
		if len(middle) > 1 {
			middle = fmt.Sprintf("(%s)%s", middle, opt)
		} else {
			middle += opt
		}
		if isPrefix {
			result += fmt.Sprintf("%s%s", substr, middle)
		} else {
			result += fmt.Sprintf("%s%s", middle, substr)
		}
		if data.end < len(vs) {
			result += "|" + ShortRegexpString(vs[data.end:]...)
		}
		return result
	}

	bestCost := 1
	for _, v := range vs {
		bestCost += len(v) + 1
	}
	best := ""

	found := false

	prefixes := commonPrefixes(vs, 1)
	for _, k := range keys(prefixes) {
		str := recurse(k, prefixes[k], true, strings.TrimPrefix)
		if len(str) < bestCost {
			bestCost = len(str)
			best = str
			found = true
		}
	}

	suffixes := commonSuffixes(vs, 1)
	for _, k := range keys(suffixes) {
		str := recurse(k, suffixes[k], false, strings.TrimSuffix)
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

func keys(m map[string]commonSs) (result []string) {
	result = make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
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

// reverseStrings is a sort.Interface that sort strings by their reverse values.
type reverseStrings []string

func (rs reverseStrings) Less(i, j int) bool {
	for m, n := len(rs[i])-1, len(rs[j])-1; m >= 0 && n >= 0; m, n = m-1, n-1 {
		if rs[i][m] != rs[j][n] {
			for ; m > 0 && !utf8.RuneStart(rs[i][m]); m-- {
			}
			for ; n > 0 && !utf8.RuneStart(rs[j][n]); n-- {
			}
			ri, _ := utf8.DecodeRuneInString(rs[i][m:])
			rj, _ := utf8.DecodeRuneInString(rs[j][n:])
			return ri < rj
		}
	}
	return len(rs[i]) < len(rs[j])
}
func (rs reverseStrings) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs reverseStrings) Len() int      { return len(rs) }

// commonSuffixes returns a map from prefixes to number of occurrences.
func commonSuffixes(vs []string, minLength int) (result map[string]commonSs) {
	sort.Sort(reverseStrings(vs))
	debugf("Reverse-sorted: %#q", vs)
	result = make(map[string]commonSs)
	for i := 0; i < len(vs)-1; i++ {
		j := i + 1
		k := 0
		for ; k < len(vs[i]) && k < len(vs[j]); k++ {
			if vs[i][len(vs[i])-k-1] != vs[j][len(vs[j])-k-1] {
				break
			}
		}
		if k < minLength {
			continue
		}
		suffix := vs[i][len(vs[i])-k:]
		if _, exists := result[suffix]; !exists {
			first := suffixStart(vs[:i], suffix)
			// debugf("suffixStart(%#q, %#q) == %v", vs[:i], suffix, first)
			// suffixEnd(vs, suffix) - first + 1
			// == suffixEnd(vs[first:], suffix) + 1
			// == suffixEnd(vs[first+1:], suffix) + 2
			end := first + 1 + suffixEnd(vs[first+1:], suffix)
			result[suffix] = commonSs{
				first, end,
			}
			debugf("suffixEnd(%#q, %#q) == %v", vs, suffix, result[suffix].end)
		}
	}
	return result
}

func suffixStart(vs []string, postfix string) int {
	debugf("suffixStart(%#q, %#q)", vs, postfix)
	if postfix == "" {
		return 0
	}
	return findFirst(vs, func(s string) bool {
		return strings.HasSuffix(s, postfix)
	})
}

func suffixEnd(vs []string, postfix string) int {
	debugf("suffixEnd(%#q, %#q)", vs, postfix)
	if postfix == "" {
		return len(vs)
	}
	return findFirst(vs, func(s string) bool {
		return !strings.HasSuffix(s, postfix)
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
	debugf("==> %d", h)
	return h
}
