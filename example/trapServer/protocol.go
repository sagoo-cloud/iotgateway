package trapServer

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/consts"
	"net"
)

type ChargeProtocol struct {
}

func (c *ChargeProtocol) Encode(args []byte) (res []byte, err error) {
	return args, nil
}

// Decode 解码 如果是TCP模式的，则需要从buffer中解析出数据，然后返回数据。否则实现自己的方法
func (c *ChargeProtocol) Decode(conn net.Conn, buffer []byte) (res []byte, err error) {

	//1，数据解析处理。。。。。

	//2，解析后，触发事件，向SagooIoT 发送数据。
	var propertieData = make(map[string]interface{})
	propertieData["XXX字段1"] = "XXX值1"
	propertieData["XXX字段2"] = "XXX值2"

	//推送数据
	out := g.Map{
		"DeviceKey":         "设备key",
		"PropertieDataList": propertieData,
	}
	event.MustFire(consts.PushAttributeDataToMQTT, out)

	return buffer, nil
}
