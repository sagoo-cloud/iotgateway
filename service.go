package iotgateway

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/lib"
	"github.com/sagoo-cloud/iotgateway/log"
	"github.com/sagoo-cloud/iotgateway/model"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"github.com/sagoo-cloud/iotgateway/vars"
	"strings"
	"time"
)

// SubscribeServiceEvent  订阅平台的服务调用，需要在有新设备接入时调用
func (gw *Gateway) SubscribeServiceEvent(deviceKey string) {
	if gw.MQTTClient == nil || !gw.MQTTClient.IsConnected() {
		log.Error("【IotGateway】SubscribeServiceEvent Error：Client has lost connection with the MQTT broker.")
		return
	}
	topic := fmt.Sprintf(serviceTopic, deviceKey)
	glog.Debugf(context.Background(), "%s 设备订阅了服务调用监听topic: %s", deviceKey, topic)
	token := gw.MQTTClient.Subscribe(topic, 1, onServiceMessage)
	if token.Error() != nil {
		glog.Debug(context.Background(), "subscribe error: ", token.Error())
	}
}

// onServiceMessage 服务调用处理
var onServiceMessage mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//忽略_reply结尾的topic
	if strings.HasSuffix(msg.Topic(), "_reply") {
		return
	}
	if msg != nil {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in safeCall:", r)
			}
		}()
		//通过监听到的topic地址获取设备标识
		deviceKey := lib.GetTopicInfo("deviceKey", msg.Topic())
		var data = mqttProtocol.ServiceCallRequest{}
		glog.Debug(context.Background(), "接收到服务下发的topic：", msg.Topic())
		glog.Debug(context.Background(), "接收到服务下发的数据：", msg.Payload())

		err := gconv.Scan(msg.Payload(), &data)
		if err != nil {
			glog.Debug(context.Background(), "解析服务功能数据出错： %s", err)
			return
		}

		//触发下发事件
		data.Params["DeviceKey"] = deviceKey

		method := strings.Split(data.Method, ".")
		var up model.UpMessage
		up.MessageID = data.Id
		up.SendTime = time.Now().UnixNano() / 1e9
		up.MethodName = method[2]
		up.Topic = msg.Topic()
		vars.UpdateUpMessageMap(deviceKey, up)
		ra, ee := vars.GetUpMessageMap(deviceKey)
		log.Debug("==222===MessageHandler===========", ra, ee)
		event.MustFire(method[2], data.Params)
	}
}
