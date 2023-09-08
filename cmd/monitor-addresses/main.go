package main

import (
	"context"
	"fmt"
	"netlinkm/internal/netlink"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, unix.SIGTERM)
	defer cancel()

	watcher := netlink.NewAddrWatcher(func(ae netlink.AddrEvent) {
		fmt.Println(ae)
	})

	if err := watcher.Start(ctx); err != nil {
		panic(err)
	}
}
