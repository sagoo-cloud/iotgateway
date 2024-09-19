package network

import "time"

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
