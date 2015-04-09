package util

import "testing"

func TestShortRegexpString(t *testing.T) {
	for _, test := range []struct {
		in  []string
		out string
	}{
		{[]string{"abc", "def"}, "abc|def"},
		{[]string{"abcf", "abde"}, "ab(de|cf)"},
		{[]string{"abcdefabcf", "abcdefabde", "abc"}, "abc(defab(de|cf))?"},
		{[]string{"css/bootstrap.css", "css/bootstrap.min.css", "css/bootstrap-theme.css", "css/bootstrap-theme.min.css"},
			`css/bootstrap((-theme(\.min)?|\.min)?.css)`},
		// `css/bootstrap(-theme)?(\.min)?.css)`},	// Ideal
		{[]string{"css/bootstrap.css", "css/bootstrap.min.css", "css/bootstrap-theme.css", "css/bootstrap-theme.min.css", "css/main.css", "css/normalize.css", "css/pygment_highlights.css", "feed.xml", "img/avatar-icon.png", "js/bootstrap.js", "js/bootstrap.min.js", "js/jquery-1.11.2.min.js", "js/main.js"},
			`css/((bootstrap(-theme(\.min)?|\.min)?|normalize|main|pygment_highlights).css)|img/avatar-icon\.png|feed\.xml|js/((bootstrap(\.min)?|main|jquery-1\.11\.2\.min).js)`},
		// `css/(bootstrap(-theme(\.min)?|\.min)?|normalize|main|pygment_highlights).css|img/avatar-icon\.png|feed\.xml|js/(bootstrap(\.min)?|main|jquery-1\.11\.2\.min).js`}, // Ideal
	} {
		if got := ShortRegexpString(test.in...); got != test.out {
			t.Errorf("expected:\n\t%#q,\ngot\n\t%#q for\n\t%v", test.out, got, test.in)
		}
	}
}

func TestCommonPrefixes(t *testing.T) {
	for _, test := range []struct {
		in  []string
		out map[string]int
	}{
		{[]string{"abc", "def"}, nil},
		{[]string{"abcf", "abde"},
			map[string]int{"ab": 2}},
		{[]string{"abcf", "abcde", "abd"},
			map[string]int{"abc": 2, "ab": 3}},
	} {
		got := commonPrefixes(test.in, 2)
		want := test.out
		if len(got) != len(want) {
			t.Errorf("expected: %v, got %v for\n\t%v", test.out, got, test.in)
			continue
		}
		for k, v := range want {
			if got[k].end-got[k].start != v {
				t.Errorf("expected: %v, got %v for\n\t%v", test.out, got, test.in)
				break
			}
		}
	}
}

func TestCommonSuffixes(t *testing.T) {
	for _, test := range []struct {
		in  []string
		out map[string]int
	}{
		{[]string{"abc", "def"}, nil},
		{[]string{"fcba", "edba"},
			map[string]int{"ba": 2}},
		{[]string{"fcba", "decba", "dba"},
			map[string]int{"cba": 2, "ba": 3}},
	} {
		// log.Print(test)
		got := commonSuffixes(test.in, 2)
		// log.Print(got)
		want := test.out
		if len(got) != len(want) {
			t.Errorf("expected: %v, got %v for\n\t%v", test.out, got, test.in)
			continue
		}
		for k, v := range want {
			if got[k].end-got[k].start != v {
				t.Errorf("expected: %v, got %v for\n\t%v", test.out, got, test.in)
				break
			}
		}
	}
}
