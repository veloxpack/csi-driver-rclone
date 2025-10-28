// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package rpcpool

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"
	"syscall"

	"storj.io/drpc"
	"storj.io/drpc/drpcmigrate"
)

func stackAnnotateAddr(addr net.Addr) (ip, port uintptr) {
	if !strings.HasPrefix(addr.Network(), "tcp") {
		return 0, 0
	}
	host, portstr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return 0, 0
	}
	if port64, err := strconv.ParseInt(portstr, 10, 64); err == nil {
		port = uintptr(port64)
	}

	// we can't afford to muck about with dns every call here. if this isn't
	// an ip address, oh well.
	if ipAddr := net.ParseIP(host); ipAddr != nil {
		if ipAddr4 := ipAddr.To4(); ipAddr4 != nil {
			ip = uintptr(binary.BigEndian.Uint32(ipAddr4))
		}
	}

	return ip, port
}

var use func(x ...uintptr)

func stackAnnotated(localip, localport, remoteip, remoteport, fd uintptr, cb func() error) error {
	err := cb()
	if use != nil {
		use(localip, localport, remoteip, remoteport, fd)
	}
	return err
}

func stackAnnotate(tr drpc.Transport, cb func() error) error {
	var localip, localport, remoteip, remoteport, fd uintptr
	if conn, ok := tr.(net.Conn); ok {
		localip, localport = stackAnnotateAddr(conn.LocalAddr())
		remoteip, remoteport = stackAnnotateAddr(conn.RemoteAddr())
		for {
			if netconn, ok := conn.(interface {
				NetConn() net.Conn
			}); ok {
				conn = netconn.NetConn()
				continue
			}
			if headerConn, ok := conn.(*drpcmigrate.HeaderConn); ok {
				conn = headerConn.Conn
				continue
			}
			break
		}
		if syscallConn, ok := conn.(interface {
			SyscallConn() (syscall.RawConn, error)
		}); ok {
			if sc, err := syscallConn.SyscallConn(); err == nil {
				_ = sc.Control(func(internalfd uintptr) {
					fd = internalfd
				})
			}
		}
	}

	return stackAnnotated(localip, localport, remoteip, remoteport, fd, cb)
}
