package mqttClient

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/conf"
	"github.com/sagoo-cloud/iotgateway/lib"
	"github.com/sagoo-cloud/iotgateway/log"
	"github.com/sagoo-cloud/iotgateway/model"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"github.com/sagoo-cloud/iotgateway/vars"
	"strings"
	"sync"
)

// 保证只有一个MQTT客户端实例的互斥锁
var singleInstanceLock sync.Mutex

// MQTT客户端单例
var client mqtt.Client

var (
	cancel context.CancelFunc
)

func init() {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	go heartbeat(ctx, vars.GatewayServerConfig.Duration)
}

// GetMQTTClient 获取MQTT客户端单例
func GetMQTTClient(cf conf.MqttConfig) (mqttClient mqtt.Client, err error) {
	log.Debug("==============config:", cf)

	singleInstanceLock.Lock()
	defer singleInstanceLock.Unlock()

	// 如果客户端已存在且已连接，直接返回现有客户端
	if client != nil && client.IsConnected() {
		return client, nil
	}
	connOpts, err := getMqttClientConfig(cf)
	if err != nil {
		return nil, err
	}
	// 创建连接
	client = mqtt.NewClient(connOpts)
	// 建立连接
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		cancel()
	}
	return client, nil
}

// Publish  向mqtt服务推送数据
func Publish(topic string, payload []byte) (err error) {
	if client == nil || !client.IsConnected() {
		fmt.Println("Client has lost connection with the MQTT broker.")
		return
	}
	pubToken := client.Publish(topic, 1, false, payload)
	return pubToken.Error()
}

// PublishData  向mqtt服务推送属性数据
func PublishData(deviceKey string, payload []byte) (err error) {
	gateWayProductKey := vars.GatewayServerConfig.ProductKey
	topic := fmt.Sprintf(propertyTopic, gateWayProductKey, deviceKey)
	log.Debug("属性上报，topic: %s", topic)
	err = Publish(topic, payload)
	if err != nil {
		log.Error("publish error: %s", err.Error())
		return
	}
	return
}

// SubscribeEvent  订阅平台的服务调用
func SubscribeEvent(deviceKey string) {
	if client == nil || !client.IsConnected() {
		log.Debug("Client has lost connection with the MQTT broker.")
		return
	}
	topic := fmt.Sprintf(serviceTopic, deviceKey)
	log.Debug("topic: %s", topic)
	token := client.Subscribe(topic, 1, onMessage)
	if token.Error() != nil {
		log.Debug("subscribe error: %s", token.Error())
	}
}

var onMessage mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//忽略_reply结尾的topic
	if strings.HasSuffix(msg.Topic(), "_reply") {
		return
	}
	if msg != nil {
		//通过监听到的topic地址获取设备标识
		deviceKey := lib.GetTopicInfo("deviceKey", msg.Topic())
		var data = mqttProtocol.ServiceCallRequest{}
		log.Debug("==111==收到服务下发的topic====", msg.Topic())
		log.Debug("====收到服务下发的信息====", msg.Payload())

		err := gconv.Scan(msg.Payload(), &data)
		if err != nil {
			log.Debug("解析服务功能数据出错： %s", err.Error())
			return
		}

		//触发下发事件
		data.Params["DeviceKey"] = deviceKey

		method := strings.Split(data.Method, ".")
		var up model.UpMessage
		up.MessageID = data.Id
		up.SendTime = gtime.Timestamp()
		up.MethodName = method[2]
		up.Topic = msg.Topic()
		vars.UpdateUpMessageMap(deviceKey, up)
		ra, ee := vars.GetUpMessageMap(deviceKey)
		log.Debug("==222===MessageHandler===========", ra, ee)
		event.MustFire(method[2], data.Params)
	}
}
