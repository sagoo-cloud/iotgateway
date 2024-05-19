package mqttClient

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/sagoo-cloud/iotgateway/log"
	"github.com/sagoo-cloud/iotgateway/vars"
)

// PublishData  向mqtt服务推送属性数据
func PublishData(deviceKey string, payload []byte) (err error) {
	gateWayProductKey := vars.GatewayServerConfig.ProductKey
	topic := fmt.Sprintf(propertyTopic, gateWayProductKey, deviceKey)
	glog.Debug(context.Background(), "属性上报，topic: %s", topic, string(payload))
	err = Publish(topic, payload)
	if err != nil {
		log.Error("publish error: %s", err.Error())
		return
	}
	return
}
