package mqttClient

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/sagoo-cloud/iotgateway/vars"
)

// PublishData  向mqtt服务推送属性数据
func PublishData(deviceKey string, payload []byte) (err error) {
	gateWayProductKey := vars.GatewayServerConfig.ProductKey
	topic := fmt.Sprintf(propertyTopic, gateWayProductKey, deviceKey)
	err = Publish(topic, payload)
	if err != nil {
		return fmt.Errorf("【IotGateway】publish error: %s", err.Error())
	}
	glog.Debugf(context.Background(), "【IotGateway】属性上报，topic: %s,推送的数据：%s", topic, string(payload))
	return
}
