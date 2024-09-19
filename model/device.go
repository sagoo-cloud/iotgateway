package model

import (
	"net"
	"time"
)

// Device 设备数据
type Device struct {
	DeviceKey    string                 // 设备唯一标识
	ClientID     string                 // 客户端ID
	OnlineStatus bool                   // 在线状态
	Conn         net.Conn               // 连接
	Metadata     map[string]interface{} // 元数据
	Info         map[string]interface{} // 设备信息
	AlarmInfo    map[string]interface{} // 报警信息
	LastActive   time.Time              // 最后活跃时间
}
