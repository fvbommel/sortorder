package rope

import (
	"io"
)

// Reader is an io.Reader that reads from a Rope.
type Reader struct {
	stack []*concat // The stack of internal nodes whose right subtrees we need to visit.
	cur   leaf      // The current leaf
	pos   int       // The position in the current leaf
}

// NewReader returns a Reader that reads from the specified Rope.
func NewReader(rope Rope) *Reader {
	// Put the leftmost path on the stack.
	var reader Reader
	if rope.node != nil {
		reader.stack = make([]*concat, 0, rope.node.depth())
		reader.pushSubtree(rope.node)
	}
	return &reader
}

func (r *Reader) pushSubtree(n node) {
	for {
		if leaf, ok := n.(leaf); ok {
			r.cur = leaf
			r.pos = 0
			return
		}
		conc := n.(*concat)
		r.stack = append(r.stack, conc)
		n = conc.Left
	}
}

func (r *Reader) nextNode() {
	r.stack = r.stack[:len(r.stack)-1]
	if len(r.stack) != 0 {
		r.pushSubtree(r.stack[len(r.stack)-1])
	}
}

// Read implements io.Reader.
func (r *Reader) Read(p []byte) (n int, err error) {
	for r.pos == len(r.cur) {
		// Done reading this node.
		r.nextNode()
		if len(r.stack) == 0 {
			// Done.
			return 0, io.EOF
		}
	}

	n = copy(p, r.cur[r.pos:])
	r.pos += n
	return n, nil
}
