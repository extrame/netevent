package netevent

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"

	"go.bug.st/serial"
)

type _reactor struct {
	udp_listeners    map[int]UdpClient
	tcp_clients      map[int]TcpClient
	unix_listeners   map[string]UnixHandler
	udp_conn         map[int]*net.UDPConn
	tcp_listeners    map[int]net.Listener
	unix_conn        map[string]*net.UnixListener
	timer            []*LaterCalling
	period_timer     []*LaterCalling
	serial_listeners map[string]SerialClient
	serial_conn      map[string]io.ReadWriteCloser
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

func (p *_reactor) DialUdp(server string, port int, udp UdpClient) {
	laddr, err := net.ResolveUDPAddr("udp", server+":"+strconv.Itoa(port))
	if err == nil {
		p.dialUdp(laddr, udp)
	} else {
		fmt.Println("resolve addr err")
		return
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
	p.tcp_clients[port] = tcp
	fmt.Println("listening on " + strconv.Itoa(port))
	var c net.Listener
	var lc net.ListenConfig
	c, err = lc.Listen(ctx, "tcp", ":"+strconv.Itoa(port))
	if err == nil {
		p.tcp_listeners[port] = c
	}
	return err
}

//function for inner-file usage
//helper function for the udp listening of ipv4 and ipv6
func (p *_reactor) dialUdp(addr *net.UDPAddr, udp UdpClient) {
	p.initReactor()
	p.udp_listeners[addr.Port] = udp
	fmt.Println("connect to" + addr.String())
	c, erl := net.DialUDP("udp", nil, addr)
	if erl != nil {
		fmt.Printf("type: %T; value: %q\n", erl, erl)
	} else {
		p.udp_conn[addr.Port] = c
	}
	transport := new(udpTransport)
	transport.setConn(c)
	udp.SetUdpTransport(transport)
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
	transport := new(udpTransport)
	transport.setConn(c)
	udp.SetUdpTransport(transport)
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
		p.tcp_clients = make(map[int]TcpClient)
	}
	if p.tcp_listeners == nil {
		p.tcp_listeners = make(map[int]net.Listener)
	}
	if p.serial_listeners == nil {
		p.serial_listeners = make(map[string]SerialClient)
	}
	if p.serial_conn == nil {
		p.serial_conn = make(map[string]io.ReadWriteCloser)
	}
}
