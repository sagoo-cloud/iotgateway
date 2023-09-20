package model

import (
	"net"
)

// Device 设备数据
type Device struct {
	DeviceKey    string
	OnlineStatus bool
	NetConn      net.Conn
	Info         map[string]interface{}
	AlarmInfo    map[string]interface{}
}
