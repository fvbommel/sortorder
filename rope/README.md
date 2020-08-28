# rope [![GoDoc](https://godoc.org/github.com/fvbommel/util/rope?status.svg)](https://godoc.org/github.com/fvbommel/util/rope)

    import "github.com/fvbommel/util/rope"

Package rope implements a "heavy-weight string", which represents very long strings more efficiently (especially when many concatenations are performed).

It may also need less memory if it contains repeated substrings, or if you use several large strings that are similar to each other.

Rope values are immutable, so each operation returns its result instead of modifying the receiver. This immutability also makes them thread-safe.
