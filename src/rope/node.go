package rope

import "io"

type (
	// The internal representation of a Rope.
	node interface {
		// A rope without subtrees is at depth 0, others at
		// max(left.depth,right.depth) + 1
		depth() depthT
		length() int64

		// Slice returns a slice of the node.
		// Precondition: start < end
		slice(start, end int64) node

		dropPrefix(start int64) node
		dropPostfix(end int64) node

		io.WriterTo

		// walkLeaves calls f on each leaf of the graph in order.
		walkLeaves(f func(leaf))
	}

	depthT byte
)

var emptyNode = node(leaf("")) // The canonical empty node.

// Helper function: returns the concatenation of the arguments.
// If lhsLength or rhsLength are <= 0, they are determined automatically if
// needed.
func conc(lhs, rhs node, lhsLength, rhsLength int64) node {
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

	if lhsLength <= 0 {
		lhsLength = lhs.length()
	}
	if rhsLength <= 0 {
		rhsLength = rhs.length()
	}
	if rhsLength > int64(^rLenT(0)) {
		// Out of range
		rhsLength = 0
	}

	return &concat{
		Left:      lhs,
		Right:     rhs,
		TreeDepth: depth + 1,
		Split:     lhsLength,
		RLen:      rLenT(rhsLength),
	}
}

// Helper function: returns the concatenation of all the arguments, in order.
// nil is interpreted as an empty string. Never returns nil.
func concMany(first node, others ...node) node {
	if first == nil {
		first = emptyNode
	}
	if len(others) == 0 {
		return first
	}
	split := len(others) / 2
	lhs := concMany(first, others[:split]...)
	rhs := concMany(others[split], others[split+1:]...)
	return conc(lhs, rhs, 0, 0)
}
