package netevent

import (
	"net"
)

type UdpClient interface {
	DatagramReceived(data []byte, addr net.Addr)
}

type UnixHandler interface {
	UnixReceived(data []byte, conn *net.UnixConn)
}

type TcpClient interface {
	DataReceived(data []byte, conn net.Conn)
}

type ErrorHandler interface {
	OnError(error)
}

type SerialClient interface {
	DataReceived(data []byte)
}
