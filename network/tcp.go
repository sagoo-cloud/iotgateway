package network

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/sagoo-cloud/iotgateway/model"
	"github.com/sagoo-cloud/iotgateway/vars"
	"io"
	"net"
	"sync"
	"time"
)

// TCPServer 结构体表示 TCP 服务器
type TCPServer struct {
	*BaseServer
	listener net.Listener
	conns    sync.Map
}

// NewTCPServer 创建一个新的 TCP 服务器实例
func NewTCPServer(options ...Option) NetworkServer {
	return &TCPServer{
		BaseServer: NewBaseServer(options...),
	}
}

// Start 启动 TCP 服务器
func (s *TCPServer) Start(ctx context.Context, addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("TCP 监听失败: %v", err)
	}

	go s.cleanupInactiveDevices(ctx)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil // 正常关闭
			}
			glog.Debugf(context.Background(), "接受 TCP 连接失败: %v", err)
			continue
		}
		go s.handleConnection(ctx, conn)
	}
}

// Stop 停止 TCP 服务器
func (s *TCPServer) Stop() error {
	if s.listener != nil {
		err := s.listener.Close()
		s.conns.Range(func(key, value interface{}) bool {
			conn := value.(net.Conn)
			conn.Close()
			return true
		})
		return err
	}
	return nil
}

// SendData 向 TCP 设备发送数据
func (s *TCPServer) SendData(device *model.Device, data interface{}, param ...string) error {
	connAny, ok := s.conns.Load(device.ClientID)
	if !ok {
		return fmt.Errorf("TCP 设备 %s 未找到", device.ClientID)
	}
	conn := connAny.(net.Conn)

	var encodedData []byte
	var err error

	if s.protocolHandler != nil {
		encodedData, err = s.protocolHandler.Encode(device, data, param...)
		if err != nil {
			return fmt.Errorf("编码数据失败: %v", err)
		}
	} else {
		encodedData = []byte(fmt.Sprintf("%v\n", data))
	}

	_, err = conn.Write(encodedData)
	return err
}

// handleConnection 处理 TCP 设备连接
func (s *TCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	clientID := conn.RemoteAddr().String()
	device := s.handleConnect(clientID, conn)
	s.conns.Store(clientID, conn)
	defer func() {
		s.handleDisconnect(device)
		s.conns.Delete(clientID)
	}()

	// 创建缓冲读取器
	var reader = bufio.NewReader(conn)
	buffer := make([]byte, 1024) // 或其他适合的缓冲区大小

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.timeout > 0 {
				if err := conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
					glog.Debugf(context.Background(), "设置读取超时失败: %v\n", err)
					return
				}
			}

			var data []byte
			// 直接读取数据
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					glog.Debugf(context.Background(), "读取错误: %v\n", err)
				}
				continue
			}
			data = buffer[:n]
			fmt.Println(fmt.Sprintf("data: %x", data))

			// 如果数据长度小于等于头部长度，则初始化协议处理器
			fmt.Println(fmt.Sprintf("data len: %d, header len: %d", n, s.packetConfig.HeaderLength))
			if n <= s.packetConfig.HeaderLength {
				s.protocolHandler.Init(device, data) // 初始化协议处理器
				if device != nil {
					device.OnlineStatus = true
					device.LastActive = time.Now() // 更新设备最后活跃时间
					if device.DeviceKey != "" {
						vars.UpdateDeviceMap(device.DeviceKey, device) // 更新到全局设备列表
					}
				}
			} else {
				if s.packetConfig.Type != NoHandling {
					var err error
					data, err = s.readPacket(reader)
					if err != nil {
						if err != io.EOF {
							glog.Debugf(context.Background(), "读取错误: %v\n", err)
						}
						continue
					}
				}

				device.LastActive = time.Now()
				resData, err := s.handleReceiveData(device, data)
				if err != nil {
					glog.Debugf(context.Background(), "处理数据错误: %v\n", err)
					continue
				}

				if resData != nil {
					if err := s.SendData(device, resData); err != nil {
						glog.Debugf(context.Background(), "发送回复失败: %v\n", err)
					}
				}
			}
		}
	}
}

// readPacket 根据配置的粘包处理方式读取数据
func (s *TCPServer) readPacket(reader io.Reader) ([]byte, error) {
	switch s.packetConfig.Type {
	case FixedLength:
		data := make([]byte, s.packetConfig.FixedLength)
		_, err := io.ReadFull(reader, data)
		return data, err
	case HeaderBodySeparate:
		headerBuf := make([]byte, s.packetConfig.HeaderLength)
		_, err := io.ReadFull(reader, headerBuf)
		if err != nil {
			return nil, err
		}
		bodyLength := binary.BigEndian.Uint32(headerBuf)
		data := make([]byte, bodyLength)
		_, err = io.ReadFull(reader, data)
		return data, err
	case Delimiter:
		return readUntilDelimiter(reader, s.packetConfig.Delimiter)
	default:
		return readUntilCRLF(reader)
	}
}

// readUntilDelimiter 读取数据直到遇到指定的分隔符
func readUntilDelimiter(reader io.Reader, delimiter string) ([]byte, error) {
	var buffer bytes.Buffer
	delimiterBytes := []byte(delimiter)
	delimiterLength := len(delimiterBytes)

	for {
		b := make([]byte, 1)
		_, err := reader.Read(b)
		if err != nil {
			return nil, err
		}

		buffer.Write(b)

		if buffer.Len() >= delimiterLength {
			if bytes.Equal(buffer.Bytes()[buffer.Len()-delimiterLength:], delimiterBytes) {
				return buffer.Bytes(), nil
			}
		}
	}
}

// readUntilCRLF 读取数据直到遇到 <CR><LF>
func readUntilCRLF(reader io.Reader) ([]byte, error) {
	var buffer bytes.Buffer
	for {
		b := make([]byte, 1)
		_, err := reader.Read(b)
		if err != nil {
			return nil, err
		}

		buffer.Write(b)

		if buffer.Len() >= 2 {
			if buffer.Bytes()[buffer.Len()-2] == '\r' && buffer.Bytes()[buffer.Len()-1] == '\n' {
				return buffer.Bytes(), nil
			}
		}
	}
}
