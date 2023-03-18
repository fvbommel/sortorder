package casefolded

import (
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/fvbommel/sortorder"
)

func TestStringSortAscii(t *testing.T) {
	want := []string{
		"ab", "ABC1",
		"abc01", "abc2",
		"aBc5", "Abc10",
		"Abc11",
	}
	got := []string{
		"aBc5", "ABC1",
		"Abc11",
		"abc01", "ab",
		"Abc10", "abc2",
	}
	sort.Sort(Natural(got))
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Error: sort failed, expected: %#q, got: %#q", want, got)
	}
}

func TestStringSortUnicode(t *testing.T) {
	// U+241A is the Kelvin symbol, which sorts as 'K' (and looks like that too).
	want := []string{
		"kl", "kLm",
		"KLM1", "klm01",
		"\u212alm2",
		"Klm10", "Klm11",
	}
	got := []string{
		"kLm", "KLM1",
		"Klm11", "klm01",
		"kl", "Klm10",
		"\u212alm2",
	}
	sort.Sort(Natural(got))
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Error: sort failed, expected: %#q, got: %#q", want, got)
	}
}

func TestNaturalLess(t *testing.T) {
	testset := []struct {
		s1, s2 string
		less   bool
	}{
		{"0", "00", true},
		{"aa", "ab", true},
		{"ab", "abc", true},
		{"abc", "ad", true},
		{"ab1", "AB2", true},
		{"AB1", "ab2", true},
		{"ab12", "abc", true},
		{"ab2a", "ab10", true},
		{"a0001", "a0000001", true},
		{"a10", "abcdefgh2", true},
		{"аб2аб", "аб10аб", true},
		{"2аб", "3аб", true},
		//
		{"a1b", "a01b", true},
		{"ab01b", "ab010b", true},
		{"a01b001", "a001b01", true},
		{"a1", "a1x", true},
		{"1ax", "1b", true},
		{"1ax", "1b", true},
		//
		{"082", "83", true},
		{"9a", "083a", true},
		// Equal if we ignore case.
		{"ab1c", "ab1c", false},
		{"aB1c", "Ab1c", false},
		{"ab1c", "ab1C", false},
		// 'k' vs 'K' vs Kelvin sign (U+212A)
		{"k", "K", false},
		{"K", "k", false},
		{"k", "\u212a", false},
		{"\u212a", "k", false},
		{"K", "\u212a", false},
		{"\u212a", "K", false},
		// Kelvin sign's spot in the alphabet.
		{"j", "\u212a", true},
		{"\u212a", "l", true},
		// Kelvin sign followed by numeric comparison
		{"\u212alm2", "Klm10", true},
		{"Klm01", "\u212alm2", true},
	}
	for _, v := range testset {
		if got := NaturalLess(v.s1, v.s2); got != v.less {
			got = NaturalLess(v.s1, v.s2)
			t.Errorf("Compared %#q to %#q: expected %v, got %v",
				v.s1, v.s2, v.less, got)
		}
		// If A < B, then B < A must be false.
		// The same cannot be said if !(A < B),
		// because A might be equal to B
		if v.less {
			if NaturalLess(v.s2, v.s1) {
				t.Errorf("Reverse-compared %#q to %#q: expected false, got true",
					v.s2, v.s1)
			}
		}
	}
}

// ToUpper, then use a regular string sort.
// As this does not perform a natural sort,
// this is not directly comparable with the other sorts.
// It is only here for a sense of scale.
func BenchmarkToUpperStdStringSort(b *testing.B) {
	set := testSet(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, list := range set {
			// Upper-casing the strings is part of the algorithm
			// when using standard sort for case-insensitive sorting.
			arr := make([]string, len(list))
			for j, str := range list {
				arr[j] = strings.ToUpper(str)
			}

			// Technically this should use a custom sorter that compares
			// arr[i] and arr[j], but swaps both in both the upper-cased
			// copy and the original array.
			sort.Strings(arr)
		}
	}
}

// ToUpperNaturalSorter is a custom sort order that compares
// upper[i] and upper[j], but swaps both upper and orig.
// Assuming upper contains the upper-cased versions of orig,
// this resembles but is not identical to, the casefolded.Natural sort order.
type ToUpperNaturalSorter struct {
	upper []string
	orig  []string
}

func (s ToUpperNaturalSorter) Len() int { return len(s.orig) }

func (s ToUpperNaturalSorter) Less(i int, j int) bool {
	return sortorder.NaturalLess(s.upper[i], s.upper[j])
}

func (s ToUpperNaturalSorter) Swap(i int, j int) {
	s.upper[i], s.upper[j] = s.upper[j], s.upper[i]
	s.orig[i], s.orig[j] = s.orig[j], s.orig[i]
}

// ToUpper, then use the "regular" Natural sort order (but swap the originals too).
// This resembles, but is not identical to, the algorithm in this package.
func BenchmarkToUpperNaturalSort(b *testing.B) {
	set := testSet(300)
	orig := make([]string, len(set[0]))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, list := range set {
			// Resetting the test set to be unsorted does not count.
			b.StopTimer()
			copy(orig, list)
			b.StartTimer()

			// Upper-casing the strings is part of the algorithm
			// when using a regular Natural order for case-insensitive sorting.
			arr := make([]string, len(orig))
			for j, str := range orig {
				arr[j] = strings.ToUpper(str)
			}

			sort.Sort(ToUpperNaturalSorter{arr, orig})
		}
	}
}

// Case-folded Natural sort order.
func BenchmarkCaseFoldedNaturalStringSort(b *testing.B) {
	set := testSet(300)
	arr := make([]string, len(set[0]))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, list := range set {
			// Resetting the test set to be unsorted does not count.
			b.StopTimer()
			copy(arr, list)
			b.StartTimer()

			sort.Sort(Natural(arr))
		}
	}
}

// ToUpper everything first, then use regular string comparison.
// Does not perform a natural sort though.
func BenchmarkToUpperStdStringLess(b *testing.B) {
	set := testSet(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Upper-casing the strings is part of the algorithm
		// when using standard '<' for case-insensitive comparisons.
		arr := make([]string, len(set[0]))
		for j, str := range set[0] {
			arr[j] = strings.ToUpper(str)
		}
		for j := range arr {
			k := (j + 1) % len(arr)
			_ = arr[j] < arr[k]
		}
	}
}

// ToUpper, then use a "regular" NaturalLess.
// This resembles, but is not identical to, the algorithm in this package.
func BenchmarkToUpperNaturalLess(b *testing.B) {
	set := testSet(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Upper-casing the strings is part of the algorithm
		// when using standard '<' for case-insensitive comparisons.
		arr := make([]string, len(set[0]))
		for j, str := range set[0] {
			arr[j] = strings.ToUpper(str)
		}
		for j := range arr {
			k := (j + 1) % len(arr)
			_ = sortorder.NaturalLess(arr[j], arr[k])
		}
	}
}

// Compare using case-folded NaturalLess()
func BenchmarkCaseFoldedNaturalLess(b *testing.B) {
	set := testSet(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range set[0] {
			k := (j + 1) % len(set[0])
			_ = NaturalLess(set[0][j], set[0][k])
		}
	}
}

// Get 1000 arrays of 10000-string-arrays (less if -short is specified).
func testSet(seed int) [][]string {
	gen := &generator{
		src: rand.New(rand.NewSource(
			int64(seed),
		)),
	}
	n := 1000
	if testing.Short() {
		n = 1
	}
	set := make([][]string, n)
	for i := range set {
		strings := make([]string, 10000)
		for idx := range strings {
			// Generate a random string
			strings[idx] = gen.NextString()
		}
		set[i] = strings
	}
	return set
}

type generator struct {
	src *rand.Rand
}

func (g *generator) NextInt(max int) int {
	return g.src.Intn(max)
}

// Gets random random-length alphanumeric string.
func (g *generator) NextString() (str string) {
	// Random-length 3-8 chars part
	strlen := g.src.Intn(6) + 3
	// Random-length 1-3 num
	numlen := g.src.Intn(3) + 1
	// Random position for num in string
	numpos := g.src.Intn(strlen + 1)
	// Generate the number
	var num string
	for i := 0; i < numlen; i++ {
		num += strconv.Itoa(g.src.Intn(10))
	}
	// Put it all together
	for i := 0; i < strlen+1; i++ {
		if i == numpos {
			str += num
		} else {
			// Generate both upper-case and lower-case variants.
			c := 'a'
			if i%2 == 0 {
				c = 'A'
			}
			str += string(c + rune(g.src.Intn(16)))
		}
	}
	return str
}
