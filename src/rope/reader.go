package rope

import (
	"io"
)

type reader struct {
	stack []*Rope // The stack of nodes. The last one is the current node.
	pos   int     // The position in the current node.
}

// NewReader returns an io.Reader that reads from the specified Rope.
func NewReader(rope *Rope) io.Reader {
	// Put the leftmost path on the stack.
	reader := reader{
		stack: make([]*Rope, 0, rope.depth+1),
	}
	reader.pushSubtree(rope)
	return &reader
}

func (r *reader) pushSubtree(node *Rope) {
	for node != nil {
		r.stack = append(r.stack, node)
		node = node.left
	}
}

func (r *reader) cur() *Rope {
	return r.stack[len(r.stack)-1]
}

func (r *reader) Read(p []byte) (n int, err error) {
	for {
		if len(r.stack) == 0 {
			return 0, io.EOF
		}
		cur := r.cur()
		if r.pos == len(cur.direct) {
			// Done reading this node.
			// Drop it from the stack and start reading its right subtree (if any)
			r.stack = r.stack[:len(r.stack)-1]
			r.pushSubtree(cur.right)
			r.pos = 0
			continue // retry
		}

		n = copy(p, cur.direct[r.pos:])
		r.pos += n
		return n, nil
	}

}
