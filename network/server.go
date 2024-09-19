package network

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/sagoo-cloud/iotgateway/model"
	"github.com/sagoo-cloud/iotgateway/vars"
	"net"
	"sync"
	"time"
)

// NetworkServer 接口定义了网络服务器的通用方法
type NetworkServer interface {
	Start(ctx context.Context, addr string) error
	Stop() error
	SendData(device *model.Device, data interface{}, param ...string) error
}

// BaseServer 结构体包含 TCP 和 UDP 服务器的共同字段
type BaseServer struct {
	devices         sync.Map
	timeout         time.Duration
	protocolHandler ProtocolHandler
	cleanupInterval time.Duration
	packetConfig    PacketConfig
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
func (s *BaseServer) handleConnect(clientID string, conn net.Conn) *model.Device {
	device := &model.Device{ClientID: clientID, OnlineStatus: true, Conn: conn, LastActive: time.Now()}
	s.devices.Store(clientID, device)
	glog.Debugf(context.Background(), "设备 %s 上线\n", clientID)
	return device
}

// getDevice 获取设备实例
func (s *BaseServer) getDevice(clientID string) *model.Device {
	if device, ok := s.devices.Load(clientID); ok {
		// 将接口类型断言为*Device类型
		if device, ok := device.(*model.Device); ok {
			return device
		}
	}
	return nil
}

// handleDisconnect 处理设备离线事件
func (s *BaseServer) handleDisconnect(device *model.Device) {
	if _, ok := s.devices.LoadAndDelete(device.ClientID); ok {
		device.OnlineStatus = false
		glog.Debugf(context.Background(), "设备 %s 离线, %s\n", device.DeviceKey, device.ClientID)
	}
}

// handleReceiveData 处理接收数据事件
func (s *BaseServer) handleReceiveData(device *model.Device, data []byte) (resData interface{}, err error) {
	if s.protocolHandler == nil {
		return nil, errors.New("未设置协议处理器")
	}
	s.protocolHandler.Init(device, data) // 初始化协议处理器
	if device != nil {
		device.OnlineStatus = true
		device.LastActive = time.Now()                 // 更新设备最后活跃时间
		vars.UpdateDeviceMap(device.DeviceKey, device) // 更新到全局设备列表
	}
	return s.protocolHandler.Decode(device, data) // 解码数据
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
				device := value.(*model.Device)
				if now.Sub(device.LastActive) > s.timeout*2 {
					s.handleDisconnect(device)
				}
				return true
			})
		}
	}
}
