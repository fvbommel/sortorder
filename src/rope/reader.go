package rope

import (
	"io"
)

type reader struct {
	stack []concat // The stack of internal nodes whose right subtrees we need to visit.
	cur   leaf     // The current leaf
	pos   int      // The position in the current leaf
}

// NewReader returns an io.Reader that reads from the specified Rope.
func NewReader(rope *Rope) io.Reader {
	// Put the leftmost path on the stack.
	reader := reader{
		stack: make([]concat, 0, rope.depth()),
	}
	reader.pushSubtree(rope.node)
	return &reader
}

func (r *reader) pushSubtree(n node) {
	for {
		if leaf, ok := n.(leaf); ok {
			r.cur = leaf
			r.pos = 0
			return
		}
		conc := n.(concat)
		n = conc.left
		r.stack = append(r.stack, conc)
	}
}

func (r *reader) nextNode() {
	r.stack = r.stack[:len(r.stack)-1]
	if len(r.stack) != 0 {
		r.pushSubtree(r.stack[len(r.stack)-1])
	}
}

func (r *reader) Read(p []byte) (n int, err error) {
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
