package mqttClient

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sagoo-cloud/iotgateway/conf"
	"sync"
)

// 保证只有一个MQTT客户端实例的互斥锁
var singleInstanceLock sync.Mutex

// MQTT客户端单例
var client mqtt.Client

var (
	cancel context.CancelFunc
)

// GetMQTTClient 获取MQTT客户端单例
func GetMQTTClient(cf conf.MqttConfig) (mqttClient mqtt.Client, err error) {
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
