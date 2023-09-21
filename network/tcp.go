package network

import (
	"context"
	"fmt"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"net"
	"sync"
)

// TcpServer 服务对象
type TcpServer struct {
	addr         string                // 监听地址
	dispatcher   *EventDispatcher      // 事件分发器
	connections  map[net.Conn]struct{} // 连接列表
	mu           sync.RWMutex          // 互斥锁
	ctx          context.Context       // 上下文
	ShutdownChan chan struct{}         // 用于通知服务器关闭的通道
}

// NewServer 创建新的 Server 对象
func NewServer(addr string) *TcpServer {
	// 创建服务
	server := &TcpServer{
		addr:         addr,
		dispatcher:   &EventDispatcher{listeners: make(map[EventType]EventListeners)},
		connections:  make(map[net.Conn]struct{}),
		ctx:          context.Background(),
		ShutdownChan: make(chan struct{}),
	}

	// 注册事件监听器
	server.dispatcher.AddEventListener(EventNewConnection, &NewConnectionListener{})
	server.dispatcher.AddEventListener(EventDataReceived, &DataReceivedListener{})
	server.dispatcher.AddEventListener(EventConnectionClosed, &ConnectionClosedListener{})

	return server
}

// Start 启动服务
func (s *TcpServer) Start(ctx context.Context, protocol mqttProtocol.Protocol) {

	s.ctx = ctx
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				ln.Close()
				return
			default:
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println("Failed to accept connection", err)
					continue
				}

				s.mu.Lock()
				s.connections[conn] = struct{}{}
				s.mu.Unlock()

				s.dispatcher.DispatchEvent(&Event{
					EventType: EventNewConnection,
					Conn:      conn,
					Protocol:  protocol,
				})

				go s.handleDataReceived(conn, protocol)
			}
		}
	}()
}

// handleDataReceived 处理数据接收
func (s *TcpServer) handleDataReceived(conn net.Conn, protocol mqttProtocol.Protocol) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		//fmt.Printf("\n 原始接收数据: %x \n", buf[:n])

		s.dispatcher.DispatchEvent(&Event{
			EventType: EventDataReceived,
			Conn:      conn,
			Data:      buf[:n],
			Protocol:  protocol,
		})
	}

	// 关闭连接事件
	s.dispatcher.DispatchEvent(&Event{
		EventType: EventConnectionClosed,
		Conn:      conn,
	})

	s.mu.Lock()
	delete(s.connections, conn)
	s.mu.Unlock()
}

// StartServer 启动服务
func StartServer(ctx context.Context, addr string, protocol mqttProtocol.Protocol) {

	server := NewServer(addr)
	server.Start(ctx, protocol)
	// 接收退出信号然后关闭服务器
	select {
	case <-server.ShutdownChan:
		fmt.Print("Shutting down ", addr)
	case <-ctx.Done():
		fmt.Print("Context Done ", addr)
	}
}
