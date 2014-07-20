package rope

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/bruth/assert"
)

func TestEmptyRope(t *testing.T) {
	for _, r := range []Rope{Rope{}, New("")} {
		assert.Equal(t, int64(0), r.Len())

		assert.Equal(t, nil, r.Bytes())
		assert.Equal(t, "", r.String())

		assert.Equal(t, "", r.DropPrefix(3).String())
		assert.Equal(t, "", r.DropPrefix(-1).String())
		assert.Equal(t, "", r.DropPostfix(3).String())
		assert.Equal(t, "", r.DropPostfix(-1).String())

		assert.Equal(t, "", r.Slice(-1, 200).String())
		assert.Equal(t, "", r.Slice(0, 1).String())

		buf := bytes.NewBuffer(nil)
		r.WriteTo(buf)
		assert.Equal(t, 0, buf.Len())
	}
}

func TestAppendRope(t *testing.T) {
	r := New("123")
	r2 := r.Append(New("456"), New("abc"), New("def"))
	assert.Equal(t, "123456abcdef", r2.String())
	assert.Equal(t, "123", r.String())
}

var treeR = New("123").Append(New("456"), New("abc")).Append(New("def"))

func testAt(t *testing.T) {
	str := treeR.String()
	length := treeR.Len()
	for i := int64(0); i < length; i++ {
		assert.Equal(t, str[i], treeR.At(i))
	}
}

func TestLen(t *testing.T) {
	assert.Equal(t, int64(0), Rope{}.Len())
	assert.Equal(t, int64(12), treeR.Len())
}

func TestString(t *testing.T) {
	assert.Equal(t, "", Rope{}.String())
	assert.Equal(t, "123456abcdef", treeR.String())
}

func TestWriteTo(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	treeR.WriteTo(buf)

	assert.Equal(t, "123456abcdef", buf.String())
}

func TestSubstring(t *testing.T) {
	defer disableCoalesce()()

	// See concat_test.go for the table used.
	for _, ss := range substrings {
		orig := Rope{ss.orig}
		got := orig.Slice(ss.start, ss.end)
		msg := fmt.Sprintf("%q[%v:%v] != %q", orig, ss.start, ss.end, got)
		assert.Equal(t, ss.want, got.node, msg)
	}
}

func TestDropPrefix(t *testing.T) {
	defer disableCoalesce()()

	// See concat_test.go for the table used.
	for _, ss := range substrings {
		if ss.end < ss.orig.length() {
			// Ignore non-suffix substrings
			continue
		}
		orig := Rope{ss.orig}
		got := orig.DropPrefix(ss.start)
		msg := fmt.Sprintf("%q[%v:] != %q", orig, ss.start, got)
		assert.Equal(t, ss.want, got.node, msg)
	}
}

func TestDropPostfix(t *testing.T) {
	defer disableCoalesce()()

	// See concat_test.go for the table used.
	for _, ss := range substrings {
		if ss.start > 0 {
			// Ignore non-prefix substrings
			continue
		}
		orig := Rope{ss.orig}
		got := orig.DropPostfix(ss.end)
		msg := fmt.Sprintf("%q[:%v] != %q", orig, ss.end, got)
		assert.Equal(t, ss.want, got.node, msg)
	}
}

func TestGoString(t *testing.T) {
	for i, format := range []string{"%v", "%#v"} {
		for _, str := range []string{"abc", "\""} {
			want := fmt.Sprintf(format, str)
			if MarkGoStringedRope && i == 1 {
				// GoStringer
				want = "/*Rope*/ " + want
			}
			assert.Equal(t, want, fmt.Sprintf(format, New(str)))
		}
	}
}
