package network

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// UDPServer 结构体表示 UDP 服务器
type UDPServer struct {
	*BaseServer
	conn *net.UDPConn
}

// NewUDPServer 创建一个新的 UDP 服务器实例
func NewUDPServer(options ...Option) NetworkServer {
	return &UDPServer{
		BaseServer: NewBaseServer(options...),
	}
}

// Start 启动 UDP 服务器
func (s *UDPServer) Start(ctx context.Context, addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("解析 UDP 地址失败: %v", err)
	}

	s.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("UDP 监听失败: %v", err)
	}

	go s.cleanupInactiveDevices(ctx)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	buffer := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if s.timeout > 0 {
				if err := s.conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
					log.Printf("设置 UDP 读取超时失败: %v\n", err)
					continue
				}
			}
			n, remoteAddr, err := s.conn.ReadFromUDP(buffer)
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil // 正常关闭
				}
				log.Printf("读取 UDP 数据失败: %v", err)
				continue
			}

			clientID := remoteAddr.String()
			device, _ := s.devices.LoadOrStore(clientID, &Device{ClientID: clientID, LastActive: time.Now()})
			device.(*Device).LastActive = time.Now()

			data := buffer[:n]
			resData, err := s.handleReceiveData(device.(*Device), data)
			if err != nil {
				log.Printf("处理数据错误: %v\n", err)
				continue
			}

			if resData != nil {
				if err := s.SendData(device.(*Device), resData); err != nil {
					log.Printf("发送回复失败: %v\n", err)
				}
			}
		}
	}
}

// Stop 停止 UDP 服务器
func (s *UDPServer) Stop() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// SendData 向 UDP 设备发送数据
func (s *UDPServer) SendData(device *Device, data interface{}, param ...string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", device.ClientID)
	if err != nil {
		return fmt.Errorf("解析 UDP 地址失败: %v", err)
	}

	var encodedData []byte
	if s.protocolHandler != nil {
		encodedData, err = s.protocolHandler.Encode(nil, data, param...)
		if err != nil {
			return fmt.Errorf("编码数据失败: %v", err)
		}
	} else {
		encodedData = []byte(fmt.Sprintf("%v", data))
	}

	_, err = s.conn.WriteToUDP(encodedData, udpAddr)
	return err
}
