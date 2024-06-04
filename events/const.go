package events

const (
	PropertySetEvent = "property" // PropertySetEvent 属性设置下发事件，SagooIoT平台下发属性设置命令时触发

	GetGatewayVersionEvent = "getGatewayVersion" // ServiceCallEvent 服务调用下发事件，SagooIoT平台下发服务调用getGatewayVersion命令时触发
	GetGatewayConfig       = "getGatewayConfig"  // ServiceCallEvent 服务调用下发事件，SagooIoT平台下发服务调用getGatewayConfig命令时触发
)
