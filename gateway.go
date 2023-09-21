package iotgateway

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/conf"
	"github.com/sagoo-cloud/iotgateway/lib"
	"github.com/sagoo-cloud/iotgateway/model"
	"github.com/sagoo-cloud/iotgateway/mqttClient"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"github.com/sagoo-cloud/iotgateway/network"
	"github.com/sagoo-cloud/iotgateway/vars"
	"github.com/sagoo-cloud/iotgateway/version"
	"log"
	"strings"
	"time"
)

const (
	propertyTopic = "/sys/%s/%s/thing/event/property/pack/post"
	serviceTopic  = "/sys/+/%s/thing/service/#"
)

type gateway struct {
	Address    string
	Version    string
	Status     string
	ctx        context.Context // 上下文
	options    *conf.GatewayConfig
	MQTTClient mqtt.Client
	Server     *network.TcpServer
	Protocol   mqttProtocol.Protocol
}

var ServerGateway *gateway

func NewGateway(options *conf.GatewayConfig, protocol mqttProtocol.Protocol) *gateway {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mqttClient, err := mqttClient.GetMQTTClient(options.MqttConfig) //初始化mqtt客户端
	tcpServer := network.NewServer(options.Server.Addr)
	if err != nil {
		log.Println("mqttClient.GetMQTTClient error:", err)
	}
	g := &gateway{
		options:    options,
		Address:    options.Server.Addr,
		ctx:        ctx,
		MQTTClient: mqttClient, // will be set later
		Server:     tcpServer,
		Protocol:   protocol,
	}
	ServerGateway = g
	g.heartbeat(context.Background(), options.Server.Duration)
	return g
}
func (g *gateway) Start() error {
	g.Server.Start(g.ctx, g.Protocol)
	log.Println("TCPServer started listening on", g.Address)
	// 接收退出信号然后关闭服务器
	select {
	case <-g.Server.ShutdownChan:
		fmt.Print("Shutting down ", g.Address)
	case <-g.ctx.Done():
		fmt.Print("Context Done ", g.Address)
	}

	g.Status = "running"

	return nil
}
func (g *gateway) heartbeat(ctx context.Context, duration time.Duration) {
	if duration == 0 {
		duration = 30
	}
	ticker := time.NewTicker(time.Second * duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//设备数量
			versionInfo := version.GetVersion()
			if versionInfo == "" || versionInfo == "0.0" {
				versionInfo = "v0.0.1"
			}
			count := vars.CountDevices()
			builder := mqttProtocol.NewGatewayBatchReqBuilder().SetId(lib.RandString(16))
			builder.SetVersion("1.0")
			builder.AddProperty("Status", 0)
			builder.AddProperty("Count", count)
			builder.AddProperty("Version", versionInfo)
			builder.SetMethod("thing.event.property.pack.post")
			data := gconv.Map(builder.Build())
			outData := gjson.New(data).MustToJson()
			topic := fmt.Sprintf(propertyTopic, vars.GatewayInfo.ProductKey, vars.GatewayInfo.DeviceKey)
			log.Println("---------网关向平台发送心跳数据-------：", outData)
			log.Println("---------topic-------：", topic)
			err := g.MQTTClient.Publish(topic, 1, false, outData)
			if err != nil {
				log.Println("publish error: ", err)
			}
		case <-ctx.Done():
			log.Println("Heartbeat stopped")
			return
		}
	}
}

// SubscribeEvent  订阅平台的服务调用
func (g *gateway) SubscribeEvent(deviceKey string) {
	if g.MQTTClient == nil || !g.MQTTClient.IsConnected() {
		fmt.Println("Client has lost connection with the MQTT broker.")
		return
	}
	topic := fmt.Sprintf(serviceTopic, deviceKey)
	log.Println("topic: ", topic)
	token := g.MQTTClient.Subscribe(topic, 1, onMessage)
	if token.Error() != nil {
		log.Println("subscribe error: ", token.Error())
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
		g.Log().Debug(context.Background(), "==111==收到服务下发的topic====", msg.Topic())
		g.Log().Debug(context.Background(), "====收到服务下发的信息====", msg.Payload())

		err := gconv.Scan(msg.Payload(), &data)
		if err != nil {
			g.Log().Debugf(context.Background(), "解析服务功能数据出错： %s", err.Error())
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
		g.Log().Debug(context.Background(), "==222===MessageHandler===========", ra, ee)
		event.MustFire(method[2], data.Params)
	}
}
