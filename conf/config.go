package conf

import (
	"time"
)

type GatewayConfig struct {
	GatewayServerConfig GatewayServerConfig `json:"server"`
	MqttConfig          MqttConfig          `json:"mqtt"`
}

type GatewayServerConfig struct {
	Name         string        `json:"name"`         // 网关服务名称,启动时显示
	Addr         string        `json:"addr"`         // 网关服务地址
	NetType      string        `json:"netType"`      // 网关服务类型
	SerUpTopic   string        `json:"serUpTopic"`   // 服务上行Topic
	SerDownTopic string        `json:"serDownTopic"` // 服务下行Topic
	Duration     time.Duration `json:"duration"`     // 网关服务心跳时长
	ProductKey   string        `json:"productKey"`   // 网关产品标识
	DeviceKey    string        `json:"deviceKey"`    // 网关实例标识
	DeviceName   string        `json:"deviceName"`   // 网关系统名称
	Description  string        `json:"description"`  // 网关系统描述
	DeviceType   string        `json:"deviceType"`   // 网关系统类型
	Manufacturer string        `json:"manufacturer"` // 网关系统厂商
	PacketConfig PacketConfig  `json:"packetConfig"`
}

// PacketHandlingType 定义了处理粘包的方法类型
type PacketHandlingType int

// PacketConfig 定义了处理粘包的配置
type PacketConfig struct {
	Type         PacketHandlingType
	FixedLength  int    // 用于 FixedLength 类型
	HeaderLength int    // 用于 HeaderBodySeparate 类型
	Delimiter    string // 用于 Delimiter 类型
}

type MqttConfig struct {
	Address               string        `json:"address"`               // mqtt服务地址
	Username              string        `json:"username"`              // mqtt服务用户名
	Password              string        `json:"password"`              // mqtt服务密码
	ClientId              string        `json:"clientId"`              // mqtt客户端标识
	ClientCertificateKey  string        `json:"clientCertificateKey"`  // mqtt客户端证书密钥
	ClientCertificateCert string        `json:"clientCertificateCert"` // mqtt客户端证书
	KeepAliveDuration     time.Duration `json:"keepAliveDuration"`     // mqtt客户端保持连接时长
	Duration              time.Duration `json:"duration"`              // mqtt客户端心跳时长
}
