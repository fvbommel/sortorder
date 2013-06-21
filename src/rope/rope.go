// Rope implements a string-like binary tree, which is a more
// efficient representation for very long strings (especially when
// many concatenations are performed).
//
// Note that it may also need considerably less memory if many of its
// substrings share common structure.
//
// Rope values are immutable, so each operation returns its result instead
// of modifying the receiver. This immutability also makes them thread-safe.
package rope

import (
	"bytes"
	"io"
)

type (
	node interface {
		// A rope without subtrees is at depth 0, others at
		// max(left.depth,right.depth) + 1
		depth() depthT
		length() int64

		// Slice returns a slice of the node.
		// Precondition: start < end
		slice(start, end int64) node

		io.WriterTo

		dropPrefix(start int64) node
		dropPostfix(end int64) node
	}

	depthT byte
)

// A value to avoid allocating for statically-known empty ropes.
var emptyNode = leaf("")        // The canonical empty node.
var emptyRope = Rope{emptyNode} // A Rope containing the empty node.

// Rope represents a non-contiguous string.
// The zero value is an empty rope.
type Rope struct {
	node // The root node of this rope. May be nil.
}

// New returns a Rope representing a given string.
func New(arg string) Rope {
	if len(arg) == 0 {
		return emptyRope
	}
	return Rope{
		node: leaf(arg),
	}
}

// Materializes the Rope as a string value.
func (r Rope) String() string {
	if r.node == nil {
		return ""
	}
	// In the trivial case, avoid allocation
	if l, ok := r.node.(leaf); ok {
		return string(l)
	}
	// The rope is not contiguous.
	return string(r.Bytes())
}

// Bytes returns the string represented by this Rope as a []byte.
func (r Rope) Bytes() []byte {
	if r.node == nil {
		return nil
	}
	buf := bytes.NewBuffer(make([]byte, 0, r.Len()))
	r.WriteTo(buf)
	return buf.Bytes()
}

// Writes the value of this Rope to the writer.
func (r Rope) WriteTo(w io.Writer) (n int64, err error) {
	if r.node == nil {
		return 0, nil // Nothing to do
	}
	return r.node.WriteTo(w)
}

// Len returns the length of the string represented by the Rope.
func (r Rope) Len() int64 {
	if r.node == nil {
		return 0
	}
	return r.node.length()
}

func concMany(first node, others ...node) node {
	if len(others) == 0 {
		return first
	}
	split := len(others) / 2
	lhs := concMany(first, others[:split]...)
	rhs := concMany(others[split], others[split+1:]...)
	return conc(lhs, rhs, 0, 0)
}

// Concat returns the Rope representing the receiver concatenated
// with the argument.
func (r Rope) Concat(rhs ...Rope) Rope {
	// Handle nil-node receiver
	for r.node == nil && len(rhs) > 0 {
		r = rhs[0]
		rhs = rhs[1:]
	}
	if len(rhs) == 0 {
		return r
	}

	list := make([]node, 0, len(rhs))
	for _, item := range rhs {
		if item.node != nil {
			list = append(list, item.node)
		}
	}
	node := concMany(r, list...)
	return Rope{node: node}
}

// DropPrefix returns a postfix of a rope, starting at index.
// It's analogous to str[start:].
func (r Rope) DropPrefix(start int64) Rope {
	if start == 0 || r.node == nil {
		return r
	}
	return Rope{
		node: r.node.dropPrefix(start),
	}
}

// DropPostfix returns the prefix of a rope ending at end.
// It's analogous to str[:end].
func (r Rope) DropPostfix(end int64) Rope {
	if r.node == nil {
		return r
	}
	return Rope{
		node: r.node.dropPrefix(end),
	}
}

// Slice returns the substring of a Rope, analogous to str[start:end].
// It is equivalent to r.DropPostfix(end).DropPrefix(start).
//
// If start >= end, start > r.Len() or end == 0, an empty Rope is returned.
func (r Rope) Slice(start, end int64) Rope {
	if start < 0 {
		start = 0
	}
	if start >= end {
		return emptyRope
	}
	return r.DropPostfix(end).DropPrefix(start)
}
