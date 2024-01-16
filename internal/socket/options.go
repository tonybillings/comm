package socket

import (
	"golang.org/x/sys/unix"
	"syscall"
)

func ControlFunc(_, _ string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err != nil {
			return
		}

		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		if err != nil {
			return
		}
	})
}
