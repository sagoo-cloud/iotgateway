package network

import (
	"context"
	"github.com/sagoo-cloud/iotgateway/log"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"net"
	"sync"
)

// TcpServer 服务对象
type TcpServer struct {
	addr        string                // 监听地址
	dispatcher  *EventDispatcher      // 事件分发器
	connections map[net.Conn]struct{} // 连接列表
	mu          sync.RWMutex          // 互斥锁
	ctx         context.Context       // 上下文
}

// NewServer 创建新的 Server 对象
func NewServer(addr string) *TcpServer {
	// 创建服务
	server := &TcpServer{
		addr:        addr,
		dispatcher:  &EventDispatcher{listeners: make(map[EventType]EventListeners)},
		connections: make(map[net.Conn]struct{}),
		ctx:         context.Background(),
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
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Debug("Failed to accept connection", err)
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
