package server

type Reader interface {
	read(*Connection, []byte) (int, error)
}

type Writer interface {
	write(*Connection, []byte) (int, error)
}

type ReadWriter interface {
	Reader
	Writer
}
