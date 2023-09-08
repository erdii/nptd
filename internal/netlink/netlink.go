// based on:
// https://stackoverflow.com/questions/36347807/how-to-monitor-ip-address-change-using-rtnetlink-socket-in-go-language
// by https://stackoverflow.com/users/2500806/khamidulla
package netlink

// https://www.man7.org/linux/man-pages/man7/rtnetlink.7.html

import (
	"context"
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"
)

type netlinkListener struct {
	fd int
	sa *syscall.SockaddrNetlink
}

// Creates and binds a netlink socket for the given rtnetlink groups
func (nl *netlinkListener) bind(groups int) error {
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_ROUTE)
	if err != nil {
		return fmt.Errorf("socket: %w", err)
	}
	nl.fd = fd

	saddr := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:    uint32(0),
		Groups: uint32(groups),
	}
	if err := syscall.Bind(fd, saddr); err != nil {
		return fmt.Errorf("bind: %s", err)
	}
	nl.sa = saddr

	return nil
}

func (l *netlinkListener) ReadMsgs(ctx context.Context) ([]syscall.NetlinkMessage, error) {
	pkt := make([]byte, 2048)

	done := make(chan struct{})
	var n int
	var err error
	go func() {
		// This blocks the thread and can't be cancelled thus it is handled in a separate goroutine.
		// FYI: Cancelling the context will leak this goroutine until the Read call returns.
		n, err = syscall.Read(l.fd, pkt)
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		// Return early if context was closed.
		return nil, ctx.Err()
	case <-done:
		// Continue after syscall returned.
	}

	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	msgs, err := syscall.ParseNetlinkMessage(pkt[:n])
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return msgs, nil
}

func (l *netlinkListener) Close() error {
	return unix.Close(l.fd)
}
