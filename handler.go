package netevent

import (
	"fmt"
	"io"
)

type UdpHandler struct {
	udptransport Transport
}

type TcpHandler struct {
	tcptransport Transport
}

type SerialHandler struct {
	rw io.ReadWriteCloser
}

func (p *UdpHandler) SetUdpTransport(transport Transport) {
	p.udptransport = transport
}

func (p *TcpHandler) SetTcpTransport(transport Transport) {
	p.tcptransport = transport
}

func (p *UdpHandler) UdpWrite(data string, addr string, port int) {
	p.udptransport.Write(data, addr, port)
}

func (p *TcpHandler) TcpWrite(data string, addr string, port int) {
	//p.transport.Write(data,addr,port)
}

func (p *SerialHandler) SerialWrite(data string) {
	if p.rw != nil {
		fmt.Println(p.rw.Write([]byte(data)))
	} else {
		fmt.Println("Unset io port for serial")
	}
}

func (p *SerialHandler) SetSerialIO(rw io.ReadWriteCloser) {
	p.rw = rw
}
