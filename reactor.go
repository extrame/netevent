package netevent

import (
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"time"

	"github.com/golang/glog"
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

func (p *_reactor) Run(ctx context.Context, fn context.CancelFunc) {
	runtime.GOMAXPROCS(len(p.udp_conn) + len(p.unix_conn))
	for port, l := range p.udp_conn {
		go handleUdpConnection(l, p.udp_listeners[port])
	}
	for addr, l := range p.unix_conn {
		go handleUnixConnection(l, p.unix_listeners[addr])
	}
	for addr, l := range p.tcp_listeners {
		go handleTcpListener(l, p.tcp_clients[addr], ctx, fn)
	}
	for dev, l := range p.serial_conn {
		go handleSerialConnection(l, p.serial_listeners[dev], ctx, fn)
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
		case <-ctx.Done():
			return
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

func handleSerialConnection(rw io.ReadWriteCloser, client SerialClient, ctx context.Context, fn context.CancelFunc) {
	go func() {
		select {
		case <-ctx.Done():
			rw.Close()
			fn()
		}
	}()
	for {
		data := make([]byte, 100)
		read_length, err := rw.Read(data)
		if err != nil { // EOF, or worse
			return
		}
		if read_length > 0 {
			panicWrapping(func() {
				client.DataReceived(data[:read_length])
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

func handleTcpListener(listener net.Listener, client TcpClient, ctx context.Context, fn context.CancelFunc) {
	var isRunning = true
	go func() {
		select {
		case <-ctx.Done():
			listener.Close()
			fn()
			isRunning = false
		}
	}()
	for {
		data := make([]byte, 1024)
		conn, err := listener.Accept()
		if err != nil {
			if ec, ok := client.(ErrorHandler); ok {
				go ec.OnError(err)
			}
			if isRunning {
				continue
			} else {
				return
			}
		}
		go func(conn net.Conn) {
			for {
				readLength, err := conn.Read(data[0:])
				if err != nil { // EOF, or worse
					if ec, ok := client.(ErrorHandler); ok {
						glog.Errorln(err)
						go ec.OnError(err)
					}
					return
				}
				if !isRunning {
					conn.Close()
					return
				}
				if readLength > 0 {
					go panicWrapping(func() {
						handleOneTcpConnect(client, data[0:readLength], conn)
					})
				}
			}
		}(conn)
	}
}

func handleOneTcpConnect(client TcpClient, data []byte, conn net.Conn) {
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
