package netlink

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
)

type AddrEvent struct {
	Addr  net.IP
	Typ   string
	Iface uint32
}

func (we AddrEvent) String() string {
	return fmt.Sprintf("%s %d %s", we.Typ, we.Iface, we.Addr)
}

type AddrCallback func(AddrEvent)

func newAddrListener() (*netlinkListener, error) {
	nl := &netlinkListener{}
	return nl, nl.bind(rtnlGroup_Addrs)
}

func NewAddrWatcher(cb AddrCallback) *watcher {
	return &watcher{
		prepare: func(w *watcher, ctx context.Context) error {
			nl, err := newAddrListener()
			if err != nil {
				return err
			}
			w.nl = nl
			return nil
		},
		teardown: func(w *watcher, ctx context.Context) error {
			return w.nl.Close()
		},
		handler: func(nm syscall.NetlinkMessage) error {
			// "Parse" message data.
			t := rtnlIfAddrMsg{
				family:    nm.Data[0],
				prefixlen: nm.Data[1],
				flags:     nm.Data[2],
				scope:     nm.Data[3],
				index:     binary.NativeEndian.Uint32(nm.Data[4 : 4+8]),
			}

			// Parse message attributes.
			attrs, err := syscall.ParseNetlinkRouteAttr(&nm)
			if err != nil {
				return fmt.Errorf("parsing netlink route attr: %w", err)
			}

			// Find message attribute with ip address.
			var addr net.IP
			for _, attr := range attrs {
				switch attr.Attr.Type {
				case syscall.IFA_ADDRESS:
					addr = attr.Value
				}
			}

			// Figure out if address was added or removed.
			var typ string
			switch nm.Header.Type {
			case syscall.RTM_NEWADDR:
				typ = "new"
			case syscall.RTM_DELADDR:
				typ = "del"
			default:
				return fmt.Errorf("%w: %d", ErrInvalidNMType, nm.Header.Type)
			}

			// Notify callback.
			cb(AddrEvent{addr, typ, t.index})
			return nil
		},
	}
}
