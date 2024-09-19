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

// SubscribeSetEvent  订阅平台的属性设置，需要在有新设备接入时调用
func (gw *Gateway) SubscribeSetEvent(deviceKey string) {
	if gw.MQTTClient == nil || !gw.MQTTClient.IsConnected() {
		log.Error("Client has lost connection with the MQTT broker.")
		return
	}
	topic := fmt.Sprintf(setTopic, deviceKey)
	glog.Debugf(context.Background(), "%s 设备订阅了属性设置监听topic: %s", deviceKey, topic)
	token := gw.MQTTClient.Subscribe(topic, 1, onSetMessage)
	if token.Error() != nil {
		glog.Debug(context.Background(), "subscribe error: ", token.Error())
	}
}

// onSetMessage 属性设置调用处理
var onSetMessage mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	if msg != nil {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in safeCall:", r)
			}
		}()
		ctx := context.Background()
		//通过监听到的topic地址获取设备标识
		deviceKey := lib.GetTopicInfo("deviceKey", msg.Topic())
		var data = mqttProtocol.ServiceCallRequest{}
		glog.Debug(ctx, "接收到属性设置下发的topic：", msg.Topic())
		glog.Debug(ctx, "接收收到属性设置下发的数据：", msg.Payload())

		err := gconv.Scan(msg.Payload(), &data)
		if err != nil {
			glog.Debug(ctx, "解析属性设置功能数据出错： %s", err)
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
		glog.Debug(ctx, "==222===MessageHandler===========", ra, ee)
		event.MustFire(method[2], data.Params)
	}
}
