package conf

import (
	"time"
)

type GatewayConfig struct {
	Server     ServerConfig `json:"server"`
	MqttConfig MqttConfig   `json:"mqtt"`
}

type ServerConfig struct {
	Addr         string        `json:"addr"`         // 网关服务地址
	Duration     time.Duration `json:"duration"`     // 网关服务心跳时长
	ProductKey   string        `json:"productKey"`   // 网关产品标识
	DeviceKey    string        `json:"deviceKey"`    // 网关实例标识
	DeviceName   string        `json:"deviceName"`   // 网关系统名称
	Description  string        `json:"description"`  // 网关系统描述
	DeviceType   string        `json:"deviceType"`   // 网关系统类型
	Manufacturer string        `json:"manufacturer"` // 网关系统厂商
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
