package rope

import (
	"io"
)

// Helper function: returns the concatenation of the arguments.
func conc(lhs, rhs node) node {
	if lhs == emptyNode {
		return rhs
	}
	if rhs == emptyNode {
		return lhs
	}

	depth := lhs.depth()
	if d := rhs.depth(); d > depth {
		depth = d
	}

	return concat{
		left:      lhs,
		right:     rhs,
		treedepth: depth + 1,
		leftLen:   lhs.length(),
	}
}

// A node representing the concatenation of two smaller nodes.
type concat struct {
	left, right node  // Subtrees. Neither may be nil or length zero.
	treedepth   int   // Depth of tree.
	leftLen     int64 // Length of left subtree.
}

func (c concat) depth() int    { return c.treedepth }
func (c concat) length() int64 { return c.leftLen + c.right.length() }

func (c concat) WriteTo(w io.Writer) (n int64, err error) {
	m, e := c.left.WriteTo(w)
	n += m
	if e != nil {
		return n, e
	}

	m, e = c.right.WriteTo(w)
	n += m
	return n, e
}

func (c concat) dropPrefix(start int64) node {
	switch {
	case start <= 0:
		return c
	case start < c.leftLen:
		return conc(c.left.dropPrefix(start), c.right)
	default: //start >= c.leftLen
		return c.right.dropPrefix(start - c.leftLen)
	}
}

func (c concat) dropPostfix(end int64) node {
	switch {
	case end <= 0:
		return emptyNode
	case end <= c.leftLen:
		return c.left.dropPostfix(end)
	default: // end > c.leftLen
		return conc(c.left, c.right.dropPostfix(end-c.leftLen))
	}
}
