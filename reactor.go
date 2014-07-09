package netevent

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"time"
)

var (
	Reactor        = new(_reactor)
	listening_chan chan int
)

type LaterCalling struct {
	millisecond int
	call        func() error
}

//type reactor interface {
//	ListenUdp(port int, client UdpClient)
//	ListenUnix(net, addr string)
//	//ListenSerial(addr string, client SerialClient, config goserial.)
//	CallLater(microsecond int, latercaller func())
//	Run()
//}

func (p *_reactor) CallLater(millisecond int, lc func() error) {
	calling := new(LaterCalling)
	calling.millisecond = millisecond
	calling.call = lc
	p.timer = append(p.timer, calling)
}

func (p *_reactor) CallPeriodly(millisecond int, lc func() error) {
	calling := new(LaterCalling)
	calling.millisecond = millisecond
	calling.call = lc
	p.period_timer = append(p.period_timer, calling)
}

func (p *_reactor) Run() {
	runtime.GOMAXPROCS(len(p.udp_conn) + len(p.unix_conn))
	for port, l := range p.udp_conn {
		go handleUdpConnection(l, p.udp_listeners[port])
	}
	for addr, l := range p.unix_conn {
		go handleUnixConnection(l, p.unix_listeners[addr])
	}
	for addr, l := range p.tcp_listeners {
		go handleTcpListener(l, p.tcp_clients[addr])
	}
	for dev, l := range p.serial_conn {
		go handleSerialConnection(l, p.serial_listeners[dev])
	}
	go func() {
		for len(p.timer) > 0 {
			caller := p.timer[0]
			p.timer = p.timer[1:]
			selectTimer(caller)
		}
	}()
	go func() {
		for len(p.period_timer) > 0 {
			caller := p.period_timer[0]
			p.period_timer = p.period_timer[1:]
			selectPeriodTimer(caller)
		}
	}()
	for {
		fmt.Println("============")
		select {
		case <-listening_chan:
			fmt.Println("--------")
		}
	}
}

func (p *_reactor) handlePeriodEvent() {

}

func selectTimer(caller *LaterCalling) {
	select {
	case <-time.After(time.Duration(caller.millisecond) * time.Millisecond):
		res := caller.call()
		if res != nil {
			fmt.Println(res)
		}
	}
}

func selectPeriodTimer(caller *LaterCalling) {
	if res := caller.call(); res == nil {
		select {
		case <-time.After(time.Duration(caller.millisecond) * time.Millisecond):
			selectPeriodTimer(caller)
		}
	} else {
		fmt.Printf("Period execuation stop by err : %s", res)
	}
}

func handleUdpConnection(conn *net.UDPConn, client UdpClient) {
	for {
		data := make([]byte, 512)
		read_length, remoteAddr, err := conn.ReadFromUDP(data[0:])
		if err != nil { // EOF, or worse
			return
		}
		if read_length > 0 {
			go panicWrapping(func() {
				client.DatagramReceived(data[0:read_length], remoteAddr)
			})
		}
	}
}

func handleSerialConnection(rw io.ReadWriteCloser, client SerialClient) {
	for {
		data := make([]byte, 1)
		read_length, err := rw.Read(data)
		if err != nil { // EOF, or worse
			return
		}
		if read_length > 0 {
			go panicWrapping(func() {
				client.DataReceived(data)
			})
		}
	}
}

func panicWrapping(f func()) {
	defer func() {
		recover()
	}()
	f()
}

func handleTcpListener(listener *net.TCPListener, client TcpClient) {
	for {
		data := make([]byte, 1024)
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		read_length, err := conn.Read(data[0:])
		if err != nil { // EOF, or worse
			fmt.Println(err)
			continue
		}
		if read_length > 0 {
			go panicWrapping(func() {
				handleOneTcpConnect(client, data[0:read_length], conn)
			})
		}
	}
}

func handleOneTcpConnect(client TcpClient, data []byte, conn *net.TCPConn) {
	defer conn.Close()
	client.DataReceived(data, conn)
}

func handleUnixConnection(listener *net.UnixListener, unix UnixHandler) {
	for {
		data := make([]byte, 512)
		conn, err := listener.AcceptUnix()
		if err != nil {
			fmt.Println(err)
			continue
		}
		read_length, err := conn.Read(data[0:])
		if err != nil { // EOF, or worse
			fmt.Println(err)
			continue
		}
		if read_length > 0 {
			go panicWrapping(func() {
				unix.UnixReceived(data[0:read_length], conn)
			})
		}
	}
}
