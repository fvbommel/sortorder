package util

import "testing"

func TestShortRegexpString(t *testing.T) {
	for _, test := range []struct {
		in  []string
		out string
	}{
		{[]string{"abc", "def"}, "abc|def"},
		{[]string{"abcf", "abde"}, "ab(cf|de)"},
		{[]string{"abcdefabcf", "abcdefabde", "abc"}, "abc(|defab(cf|de))"},
		{[]string{"css/bootstrap.css", "css/bootstrap.min.css", "css/bootstrap-theme.css", "css/bootstrap-theme.min.css", "css/main.css", "css/normalize.css", "css/pygment_highlights.css", "feed.xml", "img/avatar-icon.png", "js/bootstrap.js", "js/bootstrap.min.js", "js/jquery-1.11.2.min.js", "js/main.js"}, `css/(bootstrap(-theme.(css|min\.css)|.(css|min\.css))|main\.css|normalize\.css|pygment_highlights\.css)|feed\.xml|img/avatar-icon\.png|js/(bootstrap.(js|min\.js)|jquery-1\.11\.2\.min\.js|main\.js)`},
	} {
		if got := ShortRegexpString(test.in...); got != test.out {
			t.Errorf("expected: %#q, got %#q for\n\t%v", test.out, got, test.in)
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
