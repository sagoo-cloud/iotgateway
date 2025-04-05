package mqttClient

import (
	"context"
	"fmt"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sagoo-cloud/iotgateway/conf"
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

	// 如果客户端已存在但未连接，尝试重连
	if client != nil {
		if client.IsConnected() {
			return client, nil
		}
		// 如果存在重连管理器，尝试重连
		if reconnectManager != nil {
			if err := reconnectManager.Reconnect(); err != nil {
				return nil, fmt.Errorf("重连失败: %v", err)
			}
			return client, nil
		}
	}

	connOpts, err := getMqttClientConfig(cf)
	if err != nil {
		return nil, fmt.Errorf("failed to get MQTT client config: %v", err)
	}

	// 创建连接
	client = mqtt.NewClient(connOpts)
	if client == nil {
		return nil, fmt.Errorf("failed to create MQTT client")
	}

	// 初始化重连管理器
	reconnectManager = NewReconnectManager(client)

	// 建立连接
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		// 提供详细的错误信息
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	return client, nil
}

// Publish  向mqtt服务推送数据
func Publish(topic string, payload []byte) (err error) {
	if client == nil {
		return fmt.Errorf("【IotGateway】Publish err: client is nil. Client has lost connection with the MQTT broker.")
	}
	if !client.IsConnected() {
		// 尝试重连
		if reconnectManager != nil {
			if err := reconnectManager.Reconnect(); err != nil {
				return fmt.Errorf("【IotGateway】Publish err: 重连失败: %v", err)
			}
			// 重连成功后重试发布
			pubToken := client.Publish(topic, 1, false, payload)
			return pubToken.Error()
		}
		return fmt.Errorf("【IotGateway】Publish err: client is close. Client has lost connection with the MQTT broker.")
	}

	pubToken := client.Publish(topic, 1, false, payload)
	return pubToken.Error()
}

// GetReconnectStatus 获取重连状态
func GetReconnectStatus() map[string]interface{} {
	if reconnectManager != nil {
		return reconnectManager.GetReconnectStatus()
	}
	return nil
}

// Stop 停止MQTT客户端
func Stop() {
	if reconnectManager != nil {
		reconnectManager.Stop()
	}
	if client != nil && client.IsConnected() {
		client.Disconnect(250)
	}
}
