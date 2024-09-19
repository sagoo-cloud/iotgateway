package network

import "github.com/sagoo-cloud/iotgateway/model"

// ProtocolHandler 接口定义了协议处理方法
type ProtocolHandler interface {
	Init(device *model.Device, data []byte) error
	Encode(device *model.Device, data interface{}, param ...string) ([]byte, error)
	Decode(device *model.Device, data []byte) ([]byte, error)
}
