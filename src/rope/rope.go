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
		io.WriterTo
		dropPrefix(start int64) node
		dropPostfix(end int64) node

		// A rope without subtrees is at depth 0, others at
		// max(left.depth,right.depth) + 1
		depth() int
		length() int64
	}

	concat struct {
		left, right node  // Subtrees. Neither may be nil or length zero.
		treedepth   int   // Depth of tree.
		leftLen     int64 // Length of left subtree.
	}

	leaf string
)

// A value to avoid allocating for statically-known empty ropes.
var emptyNode = leaf("")
var emptyRope = Rope{emptyNode}

// Rope represents a non-contiguous string.
// The zero value is an empty rope.
type Rope struct {
	node // The root node of this rope. May be nil.
}

// New returns a Rope representing a given string.
func New(arg string) Rope {
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

func (l leaf) WriteTo(w io.Writer) (n int64, err error) {
	n1, err := io.WriteString(w, string(l))
	return int64(n1), err
}

func (c concat) depth() int { return c.treedepth }
func (l leaf) depth() int   { return 0 }

func (c concat) length() int64 { return c.leftLen + c.right.length() }
func (l leaf) length() int64   { return int64(len(l)) }

// Len returns the length of the string represented by the Rope.
func (r Rope) Len() int64 {
	if r.node == nil {
		return 0
	}
	return r.node.length()
}

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

func concMany(first node, others ...node) node {
	if len(others) == 0 {
		return first
	}
	split := len(others) / 2
	lhs := concMany(first, others[:split]...)
	rhs := concMany(others[split], others[split+1:]...)
	return conc(lhs, rhs)
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

func maxdepth(nodes ...node) (depth int) {
	for _, n := range nodes {
		if d := n.depth(); d > depth {
			depth = d
		}
	}
	return
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

func (l leaf) dropPrefix(start int64) node {
	switch {
	case start >= int64(len(l)):
		return emptyNode
	case start <= 0:
		return l
	default: // 0 < start < len(l)
		return l[start:]
	}
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

func (l leaf) dropPostfix(end int64) node {
	switch {
	case end >= int64(len(l)):
		return l
	case end <= 0:
		return emptyNode
	default:
		return l[:end]
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
	return r.DropPostfix(end).DropPrefix(start)
}
