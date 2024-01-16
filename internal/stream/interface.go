package stream

import "io"

// Stream Generic interface for an object that ingests bytes
// via io.Reader and outputs T on a read-only channel.  Bytes
// are read to construct new instances of T until the reader
// produces an error or the stream is explicitly closed.
type Stream[T any] interface {
	Stream(io.Reader, ...int) <-chan T
	Error() error
	io.Closer
}
