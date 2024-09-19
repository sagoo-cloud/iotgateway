package network

// ProtocolHandler 接口定义了协议处理方法
type ProtocolHandler interface {
	Init(device *Device, data []byte) error
	Encode(device *Device, data interface{}, param ...string) ([]byte, error)
	Decode(device *Device, data []byte) ([]byte, error)
}
