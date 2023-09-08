package netlink

import "syscall"

const (
	rtnlGroup_Addrs4 = (1 << (syscall.RTNLGRP_IPV4_IFADDR - 1))
	rtnlGroup_Addrs6 = (1 << (syscall.RTNLGRP_IPV6_IFADDR - 1))
	rtnlGroup_Addrs  = rtnlGroup_Addrs4 | rtnlGroup_Addrs6
)
