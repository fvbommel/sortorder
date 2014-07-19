package rope

import (
	"io"
)

// A node representing the concatenation of two smaller nodes.
type (
	concat struct {
		// Subtrees. Neither may be nil or length zero.
		left, right node
		// Length of left subtree. (relative index where the substrings meet)
		split int64
		// Cached length of right subtree, or 0 if out of range.
		rLen rLenT
		// Cached depth of the tree.
		treedepth depthT
	}

	rLenT uint32
)

func (c *concat) depth() depthT { return c.treedepth }

func (c *concat) length() int64 {
	return c.split + c.rLength()
}

func (c *concat) rLength() int64 {
	if c.rLen > 0 {
		return int64(c.rLen)
	}
	return c.right.length()
}

func (c *concat) WriteTo(w io.Writer) (n int64, err error) {
	n, err = c.left.WriteTo(w)
	if err != nil {
		return
	}

	m, err := c.right.WriteTo(w)
	return n + m, err
}

// Precondition: start < end
func (c *concat) slice(start, end int64) node {
	// If only slicing into one side, recurse to that side.
	if end <= c.split {
		return c.left.slice(start, end)
	}
	if start >= c.split {
		return c.right.slice(start-c.split, end-c.split)
	}
	clength := c.split + c.right.length()
	if start <= 0 && end >= clength {
		return c
	}

	left := c.left
	leftLen := c.split
	if start > 0 || end < c.split {
		left = left.slice(start, end)
		leftLen = -1 // Recompute if needed.
	}

	right := c.right
	rightLen := int64(c.rLen)
	if start > c.split || end < clength {
		right = c.right.slice(start-c.split, end-c.split)
		rightLen = -1 // Recompute if needed.
	}

	return conc(left, right, leftLen, rightLen)
}

func (c *concat) dropPrefix(start int64) node {
	switch {
	case start <= 0:
		return c
	case start < c.split:
		return conc(c.left.dropPrefix(start), c.right,
			c.split-start, int64(c.rLen))
	default: // start >= c.split
		return c.right.dropPrefix(start - c.split)
	}
}

func (c *concat) dropPostfix(end int64) node {
	switch {
	case end <= 0:
		return emptyNode
	case end <= c.split:
		return c.left.dropPostfix(end)
	case end >= c.split+c.rLength():
		return c
	default: // c.split < end < c.length()
		end -= c.split
		return conc(c.left, c.right.dropPostfix(end), c.split, end)
	}
}

func (c *concat) walkLeaves(f func(leaf)) {
	c.left.walkLeaves(f)
	c.right.walkLeaves(f)
}
