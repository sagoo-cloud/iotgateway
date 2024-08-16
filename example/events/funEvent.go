package events

import (
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/events"
)

func Init() {
	// 向设备发送命令
	event.On(events.PropertySetEvent, event.ListenerFunc(modbusWriteData), event.Normal)

	//如果要实现服务上发事件，可以在这里添加，事件名称为SagooIoT中定义的下发服务标识名称
}

// modbusWriteData 向设备发送命令事件
func modbusWriteData(e event.Event) error {
	//在这儿里实现，向设备下发的处理
	return nil
}
