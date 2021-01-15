package netevent

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"go.bug.st/serial"
)

type _reactor struct {
	udp_listeners map[int]UdpClient
	udp_conn      map[int]*net.UDPConn

	tcp_clients   map[string]TcpClient
	tcp_listeners map[string]net.Listener
	tcp_conn      map[string]*net.TCPConn

	unix_listeners map[string]UnixHandler
	unix_conn      map[string]*net.UnixListener

	serial_listeners map[string]SerialClient
	serial_conn      map[string]io.ReadWriteCloser

	timer        []*LaterCalling
	period_timer []*LaterCalling
}

func (p *_reactor) ListenUnix(addr string, unix UnixHandler) {
	fmt.Printf("Add listener on %s\n", addr)
	if p.unix_listeners == nil {
		p.unix_listeners = make(map[string]UnixHandler)
	}
	if p.unix_conn == nil {
		p.unix_conn = make(map[string]*net.UnixListener)
	}
	p.unix_listeners[addr] = unix
	laddr, err := net.ResolveUnixAddr("unix", addr)
	if err != nil {
		fmt.Println("resolve addr err")
		return
	}
	c, erl := net.ListenUnix("unix", laddr)
	if erl != nil {
		fmt.Printf("Listen err, type: %T; value: %q\n", erl, erl)
	} else {
		p.unix_conn[addr] = c
	}
}

func (p *_reactor) DialUdp(server string, port int, udp UdpClient) (*net.UDPConn, error) {
	laddr, err := net.ResolveUDPAddr("udp", server+":"+strconv.Itoa(port))
	if err == nil {
		return p.dialUdp(laddr, udp)
	} else {
		log.Println("resolve addr err")
		return nil, err
	}
}

func (p *_reactor) DialTcp(server string, port int, tcp TcpClient) (*net.TCPConn, error) {
	laddr, err := net.ResolveTCPAddr("tcp", server+":"+strconv.Itoa(port))
	if err == nil {
		return p.dialTcp(laddr, tcp)
	} else {
		log.Println("resolve addr err")
		return nil, err
	}
}

func (p *_reactor) ListenUdp(port int, udp UdpClient) {
	laddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	if err == nil {
		p.listenUdp(laddr, udp)
	} else {
		fmt.Println("resolve addr err")
		return
	}
}

func (p *_reactor) ListenTCP(ctx context.Context, port int, tcp TcpClient) (err error) {
	p.initReactor()
	addr := ":" + strconv.Itoa(port)
	p.tcp_clients[addr] = tcp
	fmt.Println("listening on " + addr)
	var c net.Listener
	var lc net.ListenConfig
	c, err = lc.Listen(ctx, "tcp", addr)
	if err == nil {
		p.tcp_listeners[addr] = c
	}
	return err
}

//function for inner-file usage
//helper function for the udp listening of ipv4 and ipv6
func (p *_reactor) dialUdp(addr *net.UDPAddr, udp UdpClient) (c *net.UDPConn, err error) {
	p.initReactor()
	p.udp_listeners[addr.Port] = udp
	fmt.Println("connect to" + addr.String())
	c, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("type: %T; value: %q\n", err, err)
	} else {
		p.udp_conn[addr.Port] = c
	}
	return
}

//function for inner-file usage
//helper function for the udp listening of ipv4 and ipv6
func (p *_reactor) dialTcp(addr *net.TCPAddr, tcp TcpClient) (c *net.TCPConn, err error) {
	p.initReactor()
	p.tcp_clients[addr.String()] = tcp
	fmt.Println("connect to" + addr.String())
	c, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Printf("type: %T; value: %q\n", err, err)
	} else {
		p.tcp_conn[addr.String()] = c
	}
	return
}

//function for inner-file usage
//helper function for the udp listening of ipv4 and ipv6
func (p *_reactor) listenUdp(addr *net.UDPAddr, udp UdpClient) {
	p.initReactor()
	p.udp_listeners[addr.Port] = udp
	fmt.Println("listening on " + strconv.Itoa(addr.Port))
	c, erl := net.ListenUDP("udp", addr)
	if erl != nil {
		fmt.Printf("type: %T; value: %q\n", erl, erl)
	} else {
		p.udp_conn[addr.Port] = c
	}
}

func (p *_reactor) ListenSerial(dev string, client SerialClient, baud int) (rw io.ReadWriteCloser, err error) {
	p.initReactor()
	p.serial_listeners[dev] = client
	fmt.Printf("listening on %s with (%d) Bund\n", dev, baud)
	var s *serial.SerialPort
	s, err = serial.OpenPort(dev, &serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		Parity:   serial.PARITY_NONE,
		StopBits: serial.STOPBITS_ONE,
	})
	if err == nil {
		p.serial_conn[dev] = s
	}
	return s, err
}

//init the reactor
func (p *_reactor) initReactor() {
	if p.udp_listeners == nil {
		p.udp_listeners = make(map[int]UdpClient)
	}
	if p.udp_conn == nil {
		p.udp_conn = make(map[int]*net.UDPConn)
	}
	if p.tcp_clients == nil {
		p.tcp_clients = make(map[string]TcpClient)
	}
	if p.tcp_listeners == nil {
		p.tcp_listeners = make(map[string]net.Listener)
	}
	if p.tcp_conn == nil {
		p.tcp_conn = make(map[string]*net.TCPConn)
	}
	if p.serial_listeners == nil {
		p.serial_listeners = make(map[string]SerialClient)
	}
	if p.serial_conn == nil {
		p.serial_conn = make(map[string]io.ReadWriteCloser)
	}
}
