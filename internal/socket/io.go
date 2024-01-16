package socket

import (
	"golang.org/x/sys/unix"
	"os"
)

func SinkReadWriteError(err error) error {
	switch err {
	case nil, unix.EINTR, unix.EAGAIN, unix.EINPROGRESS:
		return nil
	default:
		if os.IsTimeout(err) {
			err = nil
		}
		return err
	}
}
