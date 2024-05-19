package iotgateway

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/sagoo-cloud/iotgateway/conf"
	"github.com/sagoo-cloud/iotgateway/consts"
	"github.com/sagoo-cloud/iotgateway/log"
	"github.com/sagoo-cloud/iotgateway/mqttClient"
	"github.com/sagoo-cloud/iotgateway/mqttProtocol"
	"github.com/sagoo-cloud/iotgateway/network"
	"github.com/sagoo-cloud/iotgateway/vars"
	"github.com/sagoo-cloud/iotgateway/version"
	"strings"
	"time"
)

const (
	propertyTopic = "/sys/%s/%s/thing/event/property/pack/post"
	serviceTopic  = "/sys/+/%s/thing/service/#"
	setTopic      = "/sys/+/%s/thing/service/property/set"
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
	cancel     context.CancelFunc
}

var ServerGateway *gateway

func NewGateway(options *conf.GatewayConfig, protocol mqttProtocol.Protocol) (gw *gateway, err error) {
	client, err := mqttClient.GetMQTTClient(options.MqttConfig) //初始化mqtt客户端
	if err != nil {
		log.Debug("mqttClient.GetMQTTClient error:", err)
	}
	if options.GatewayServerConfig.NetType == "" {
		options.GatewayServerConfig.NetType = consts.NetTypeTcpServer
	}
	vars.GatewayServerConfig = options.GatewayServerConfig

	gw = &gateway{
		options:    options,
		Address:    options.GatewayServerConfig.Addr,
		MQTTClient: client, // will be set later
		Server:     nil,
		Protocol:   protocol,
	}
	gw.ctx, gw.cancel = context.WithCancel(context.Background())
	defer gw.cancel()
	ServerGateway = gw
	return
}
func (gw *gateway) Start() {
	name := gw.options.GatewayServerConfig.Name
	if name == "" {
		name = "SagooIoT Gateway Server"
	}
	go gw.heartbeat(gw.options.GatewayServerConfig.Duration) //启动心跳
	switch gw.options.GatewayServerConfig.NetType {
	case consts.NetTypeTcpServer:
		//启动tcp类型的设备网关服务
		log.Info("%s started listening on %s", name, gw.Address)
		gw.Server = network.NewServer(gw.options.GatewayServerConfig.Addr)
		gw.Server.Start(gw.ctx, gw.Protocol)
	case consts.NetTypeMqttServer:
		//启动mqtt类型的设备网关服务
		glog.Infof(context.Background(), "%s started listening ......", name)
		//log.Info("%s started listening ......")
		gw.SubscribeDeviceUpData()
		select {}
	}

	return
}

// heartbeat 网关服务心跳
func (gw *gateway) heartbeat(duration time.Duration) {
	if duration == 0 {
		duration = 60
	}
	ticker := time.NewTicker(time.Second * duration)
	//defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//设备数量
			versionInfo := version.GetVersion()
			if versionInfo == "" || versionInfo == "0.0" {
				versionInfo = "v0.0.1"
			}
			count := vars.CountDevices()
			builder := mqttProtocol.NewGatewayBatchReqBuilder().SetId(guid.S())
			builder.SetVersion("1.0")
			builder.AddProperty("Status", 0)
			builder.AddProperty("Count", count)
			builder.AddProperty("Version", versionInfo)
			builder.SetMethod("thing.event.property.pack.post")
			data := gconv.Map(builder.Build())
			outData := gjson.New(data).MustToJson()
			topic := fmt.Sprintf(propertyTopic, vars.GatewayServerConfig.ProductKey, vars.GatewayServerConfig.DeviceKey)
			glog.Debugf(context.Background(), "网关向平台发送心跳数据：%s", string(outData))
			token := gw.MQTTClient.Publish(topic, 1, false, outData)
			if token.Error() != nil {
				log.Error("publish error: %s", token.Error())
			}
		}
	}
}

// SubscribeDeviceUpData 在mqtt网络类型的设备情况下，订阅设备上传数据
func (gw *gateway) SubscribeDeviceUpData() {
	if gw.MQTTClient == nil || !gw.MQTTClient.IsConnected() {
		log.Error("Client has lost connection with the MQTT broker.")
		return
	}
	log.Debug("订阅设备上传数据topic: ", gw.options.GatewayServerConfig.SerUpTopic)
	if gw.options.GatewayServerConfig.SerUpTopic != "" {
		token := gw.MQTTClient.Subscribe(gw.options.GatewayServerConfig.SerUpTopic, 1, onDeviceUpDataMessage)
		if token.Error() != nil {
			log.Debug("subscribe error: ", token.Error())
		}
	}

}

// onDeviceUpDataMessage 设备上传数据
var onDeviceUpDataMessage mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//忽略_reply结尾的topic
	if strings.HasSuffix(msg.Topic(), "_reply") {
		return
	}
	if msg != nil {
		ServerGateway.Protocol.Decode(nil, msg.Payload())
	}
}

// DeviceDownData 在mqtt网络类型的设备情况下，向设备下发数据
func (gw *gateway) DeviceDownData(data interface{}) {
	if gw.MQTTClient == nil || !gw.MQTTClient.IsConnected() {
		log.Error("Client has lost connection with the MQTT broker.")
		return
	}
	if gw.options.GatewayServerConfig.SerDownTopic != "" {
		token := gw.MQTTClient.Publish(gw.options.GatewayServerConfig.SerDownTopic, 1, false, data)
		if token.Error() != nil {
			log.Error("publish error: %s", token.Error())
		}
	}
}
