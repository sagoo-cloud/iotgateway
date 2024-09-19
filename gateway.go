package iotgateway

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/gookit/event"
	"github.com/sagoo-cloud/iotgateway/conf"
	"github.com/sagoo-cloud/iotgateway/consts"
	"github.com/sagoo-cloud/iotgateway/events"
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

type Gateway struct {
	Address    string
	Version    string
	Status     string
	ctx        context.Context // 上下文
	options    *conf.GatewayConfig
	MQTTClient mqtt.Client
	Server     network.NetworkServer
	Protocol   network.ProtocolHandler
	cancel     context.CancelFunc
}

var ServerGateway *Gateway

func NewGateway(ctx context.Context, protocol network.ProtocolHandler) (gw *Gateway, err error) {

	//读取配置文件
	options := new(conf.GatewayConfig)
	confData, err := g.Cfg().Data(ctx)
	if err != nil {
		glog.Debug(context.Background(), "读取配置文件失败", err)
		return
	}
	err = gconv.Scan(confData, options)
	if err != nil {
		glog.Error(ctx, "读取配置文件失败", err)
		return
	}

	client, err := mqttClient.GetMQTTClient(options.MqttConfig) //初始化mqtt客户端
	if err != nil {
		log.Debug("mqttClient.GetMQTTClient error:", err)
	}
	if options.GatewayServerConfig.NetType == "" {
		options.GatewayServerConfig.NetType = consts.NetTypeTcpServer
	}
	vars.GatewayServerConfig = options.GatewayServerConfig

	gw = &Gateway{
		options:    options,
		Address:    options.GatewayServerConfig.Addr,
		MQTTClient: client, // will be set later
		Server:     nil,
		Protocol:   protocol,
	}
	gw.ctx, gw.cancel = context.WithCancel(context.Background())
	defer gw.cancel()

	//初始化事件
	defer func() {
		err := event.CloseWait()
		if err != nil {
			glog.Debugf(context.Background(), "event.CloseWait() error: %s", err.Error())
		}
	}()
	events.LoadingPublishEvent() //加载发布事件

	ServerGateway = gw
	return
}
func (gw *Gateway) Start() {
	name := gw.options.GatewayServerConfig.Name
	if name == "" {
		name = "SagooIoT Gateway Server"
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//订阅网关设备服务下发事件
	gw.SubscribeServiceEvent(gw.options.GatewayServerConfig.DeviceKey)

	go gw.heartbeat(gw.options.GatewayServerConfig.Duration) //启动心跳
	switch gw.options.GatewayServerConfig.NetType {
	case consts.NetTypeTcpServer:
		// 创建 TCP 服务器
		gw.Server = network.NewTCPServer(
			network.WithTimeout(1*time.Minute),
			network.WithProtocolHandler(gw.Protocol),
			network.WithCleanupInterval(5*time.Minute),
			network.WithPacketHandling(network.PacketConfig{Type: network.Delimiter, Delimiter: "\r\n"}),
		)
		glog.Infof(ctx, "%s started Tcp listening on %v", name, gw.options.GatewayServerConfig.Addr)
		// 启动 TCP 服务器
		if err := gw.Server.Start(ctx, gw.options.GatewayServerConfig.Addr); err != nil {
			log.Info("TCP 服务器错误: %v", err)
		}

	case consts.NetTypeUDPServer:
		// 创建 UDP 服务器
		gw.Server = network.NewUDPServer(
			network.WithTimeout(1*time.Minute),
			network.WithProtocolHandler(gw.Protocol),
			network.WithCleanupInterval(5*time.Minute),
		)
		glog.Infof(ctx, "%s started UDP listening on %v", name, gw.options.GatewayServerConfig.Addr)
		// 启动 UDP 服务器
		if err := gw.Server.Start(ctx, gw.options.GatewayServerConfig.Addr); err != nil {
			log.Info("UDP 服务器错误: %v", err)
		}
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
func (gw *Gateway) heartbeat(duration time.Duration) {
	if duration == 0 {
		duration = 60
	}
	ticker := time.NewTicker(time.Second * duration)

	// 立即发送一次心跳消息
	gw.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			// 发送心跳消息
			gw.sendHeartbeat()
		}
	}
}

// sendHeartbeat 发送心跳消息
func (gw *Gateway) sendHeartbeat() {
	// 设备数量
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
		glog.Errorf(context.Background(), "publish error: %s", token.Error())
	}
}

// SubscribeDeviceUpData 在mqtt网络类型的设备情况下，订阅设备上传数据
func (gw *Gateway) SubscribeDeviceUpData() {
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
func (gw *Gateway) DeviceDownData(data interface{}) {
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
