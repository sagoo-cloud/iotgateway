package network

// PacketHandlingType 定义了处理粘包的方法类型
type PacketHandlingType int

const (
	NoHandling PacketHandlingType = iota
	FixedLength
	HeaderBodySeparate
	Delimiter
)

// PacketConfig 定义了处理粘包的配置
type PacketConfig struct {
	Type         PacketHandlingType
	FixedLength  int    // 用于 FixedLength 类型
	HeaderLength int    // 用于 HeaderBodySeparate 类型
	Delimiter    string // 用于 Delimiter 类型
}
