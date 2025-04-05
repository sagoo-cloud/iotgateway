package mqttClient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sagoo-cloud/iotgateway/conf"
	"github.com/sagoo-cloud/iotgateway/log"
)

const (
	propertyTopic = "/sys/%s/%s/thing/event/property/pack/post"
	serviceTopic  = "/sys/+/%s/thing/service/#"
)

var (
	reconnectManager *ReconnectManager
)

// getMqttClientConfig 获取mqtt客户端配置
func getMqttClientConfig(cf conf.MqttConfig) (connOpts *mqtt.ClientOptions, err error) {
	if cf.Address == "" && cf.Username == "" && cf.Password == "" && cf.ClientId == "" {
		return nil, fmt.Errorf("mqtt配置信息不完整")
	}

	connOpts = mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", cf.Address))
	connOpts.SetUsername(cf.Username)
	connOpts.SetPassword(cf.Password)
	connOpts.SetClientID(cf.ClientId)

	if cf.ClientCertificateKey != "" {
		connOpts.AddBroker(fmt.Sprintf("ssl://%s", cf.Address))
		tlsConfig := NewTlsConfig(cf.ClientCertificateKey, cf.ClientCertificateCert)
		connOpts.SetTLSConfig(tlsConfig)
	}

	connOpts.SetKeepAlive(cf.KeepAliveDuration * time.Second)
	connOpts.SetConnectTimeout(time.Second * 10)
	connOpts.SetMaxReconnectInterval(time.Minute * 5)
	connOpts.SetAutoReconnect(true)
	connOpts.SetConnectRetryInterval(time.Second * 5)

	// 连接成功回调
	connOpts.OnConnect = func(client mqtt.Client) {
		log.Debug("MQTT服务连接成功")
		if reconnectManager != nil {
			reconnectManager.currentRetry = 0
			// 连接成功后，启动重连循环以确保连接持续
			reconnectManager.StartReconnectLoop()
		}
	}

	// 连接断开回调
	connOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Debug("MQTT服务连接已断开: %v", err)
		if reconnectManager != nil {
			// 连接断开时，确保重连循环正在运行
			reconnectManager.StartReconnectLoop()
		}
	}

	// 重连回调
	connOpts.OnReconnecting = func(client mqtt.Client, o *mqtt.ClientOptions) {
		log.Debug("正在尝试重新连接MQTT服务...")
	}

	return
}

// NewTlsConfig 生成tls配置
func NewTlsConfig(clientCertificateKey, clientCertificateCert string) *tls.Config {
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile("ca.pem")
	if err != nil {
		log.Error(err.Error())
	}
	certPool.AppendCertsFromPEM(ca)
	if clientCertificateKey != "" && clientCertificateCert != "" {
		clientKeyPair, err := tls.LoadX509KeyPair(clientCertificateCert, clientCertificateKey)
		if err != nil {
			panic(err)
		}
		return &tls.Config{
			RootCAs:            certPool,
			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{clientKeyPair},
		}
	}
	return &tls.Config{
		RootCAs: certPool,
	}
}
