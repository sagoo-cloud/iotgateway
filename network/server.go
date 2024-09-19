package network

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/os/glog"
	"net"
	"sync"
	"time"
)

// Device 结构体表示一个连接的设备
type Device struct {
	DeviceKey  string                 // 设备唯一标识
	ClientID   string                 // 客户端ID
	Conn       net.Conn               // 连接
	Metadata   map[string]interface{} // 元数据
	LastActive time.Time              // 最后活跃时间
}

// NetworkServer 接口定义了网络服务器的通用方法
type NetworkServer interface {
	Start(ctx context.Context, addr string) error
	Stop() error
	SendData(device *Device, data interface{}, param ...string) error
}

// BaseServer 结构体包含 TCP 和 UDP 服务器的共同字段
type BaseServer struct {
	devices         sync.Map
	timeout         time.Duration
	protocolHandler ProtocolHandler
	cleanupInterval time.Duration
	packetConfig    PacketConfig
}

// Option 定义了服务器配置的选项函数类型
type Option func(*BaseServer)

// WithTimeout 设置超时选项
func WithTimeout(timeout time.Duration) Option {
	return func(s *BaseServer) {
		s.timeout = timeout
	}
}

// WithProtocolHandler 设置协议处理器选项
func WithProtocolHandler(handler ProtocolHandler) Option {
	return func(s *BaseServer) {
		s.protocolHandler = handler
	}
}

// WithCleanupInterval 设置清理间隔选项
func WithCleanupInterval(interval time.Duration) Option {
	return func(s *BaseServer) {
		s.cleanupInterval = interval
	}
}

// WithPacketHandling 设置粘包处理选项
func WithPacketHandling(config PacketConfig) Option {
	return func(s *BaseServer) {
		s.packetConfig = config
	}
}

// NewBaseServer 创建一个新的基础服务器实例
func NewBaseServer(options ...Option) *BaseServer {
	s := &BaseServer{
		timeout:         30 * time.Second,
		cleanupInterval: 5 * time.Minute,
		packetConfig:    PacketConfig{Type: Delimiter, Delimiter: "\r\n"},
	}

	for _, option := range options {
		option(s)
	}

	return s
}

// handleConnect 处理设备上线事件
func (s *BaseServer) handleConnect(clientID string, conn net.Conn) *Device {
	device := &Device{ClientID: clientID, Conn: conn, LastActive: time.Now()}
	s.devices.Store(clientID, device)
	glog.Debugf(context.Background(), "设备 %s 上线\n", clientID)
	return device
}

// getDevice 获取设备实例
func (s *BaseServer) getDevice(clientID string) *Device {
	if device, ok := s.devices.Load(clientID); ok {
		// 将接口类型断言为*Device类型
		if device, ok := device.(*Device); ok {
			return device
		}
	}
	return nil
}

// handleDisconnect 处理设备离线事件
func (s *BaseServer) handleDisconnect(clientID string) {
	if _, ok := s.devices.LoadAndDelete(clientID); ok {
		glog.Debugf(context.Background(), "设备 %s 离线\n", clientID)
	}
}

// handleReceiveData 处理接收数据事件
func (s *BaseServer) handleReceiveData(device *Device, data []byte) (resData interface{}, err error) {
	if s.protocolHandler == nil {
		return nil, errors.New("未设置协议处理器")
	}
	s.protocolHandler.Init(device, data)
	return s.protocolHandler.Decode(device, data)
}

// cleanupInactiveDevices 清理不活跃的设备
func (s *BaseServer) cleanupInactiveDevices(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			s.devices.Range(func(key, value interface{}) bool {
				device := value.(*Device)
				if now.Sub(device.LastActive) > s.timeout*2 {
					s.handleDisconnect(device.ClientID)
				}
				return true
			})
		}
	}
}
