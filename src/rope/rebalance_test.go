package rope

import (
	"testing"

	"github.com/bruth/assert"
	"github.com/kr/pretty"
)

var rebalanceTestRopes = []Rope{
	New("a").Append(New("bc").Append(New("d").Append(New("ef")))), //.Append(New("g")),
	New("a").Append(New("bc").Append(New("d").Append(New("ef")))).Append(New("g")),
	New("a").Append(New("bcd")).Append(New("efghijkl")).Append(New("mno")),
	New("abc").Append(New("def")).Append(New("def")).Append(New("def")).Append(New("def")).Append(New("def")).Append(New("ghiklmnopqrstuvwxyzklmnopqrstuvwxyz")).Append(New("j")).Append(New("j")).Append(New("j")).Append(New("j")).Append(New("j")).Append(New("klmnopqrstuvwxyz")),
}

func init() {
	//~ for i, r := range largeRopes {
	//~ for j := 0; j < 8; j++ {
	//~ r = r.Append(r)
	//~ }
	//~ largeRopes[i] = r
	//~ }
	//~ n := New("a")
	var r Rope
	for i := 0; i < 100; i++ {
		r = r.Append(New(string(' ' + i)))
	}
	rebalanceTestRopes = append(rebalanceTestRopes, r)

	rebalanceTestRopes = append(rebalanceTestRopes, emptyRope, Rope{})
}

func TestRebalance(t *testing.T) {
	for _, orig := range rebalanceTestRopes {
		origStr := orig.String()
		rebalanced := orig.Rebalance()
		rebalancedStr := rebalanced.String()

		pretty.Println(orig, "(", orig.isBalanced(), ") ==> (", rebalanced.isBalanced(), ")", rebalanced)

		assert.Equal(t, origStr, rebalancedStr)
		assert.True(t, rebalanced.isBalanced())
	}
}
