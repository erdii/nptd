package netlink

import (
	"context"
	"errors"
	"fmt"
	"syscall"
)

// https://www.man7.org/linux/man-pages/man7/rtnetlink.7.html

var ErrInvalidNMType = errors.New("invalid netlink message type")

type rtnlIfAddrMsg struct {
	family    byte   /* Address type */
	prefixlen byte   /* Prefixlength of address */
	flags     byte   /* Address flags */
	scope     byte   /* Address scope */
	index     uint32 /* Interface index */
}

type watcher struct {
	handler           msgHandler
	nl                *netlinkListener
	prepare, teardown func(*watcher, context.Context) error
}

type msgHandler func(syscall.NetlinkMessage) error

func (w *watcher) Start(ctx context.Context) (err error) {
	// Run preparation function to setup netlink listener.
	if err := w.prepare(w, ctx); err != nil {
		return err
	}

	// Defer teardown function for cleanup.
	// Also handle case where teardown errors while already returning an error.
	defer func() {
		if tErr := w.teardown(w, ctx); tErr != nil {
			if err != nil {
				err = fmt.Errorf("tearing down: %w, while handling another error: %w", tErr, err)
			}
			err = tErr
		}
	}()

	for {
		// Check if context has been cancelled and return early.
		if ctx.Err() != nil {
			return nil
		}

		// Read netlink messages from socket.
		var msgs []syscall.NetlinkMessage
		msgs, err = w.nl.ReadMsgs(ctx)
		if err != nil {
			return fmt.Errorf("reading netlink msg: %w", err)
		}

		// Run handler for every message.
		for _, m := range msgs {
			err := w.handler(m)
			if err != nil {
				return err
			}
		}
	}
}
