package network

import "github.com/sagoo-cloud/iotgateway/conf"

const (
	NoHandling         conf.PacketHandlingType = iota
	FixedLength                                // 定长
	HeaderBodySeparate                         // 头部+体
	Delimiter                                  // 分隔符
)
