package mqttClient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sagoo-cloud/iotgateway/conf"
	"github.com/sagoo-cloud/iotgateway/lib"
	"github.com/sagoo-cloud/iotgateway/log"
	"os"
	"time"
)

const (
	propertyTopic = "/sys/%s/%s/thing/event/property/pack/post"
	serviceTopic  = "/sys/+/%s/thing/service/#"
)

// getMqttClientConfig 获取mqtt客户端配置
func getMqttClientConfig(cf conf.MqttConfig) (connOpts *mqtt.ClientOptions, err error) {
	if cf.Address == "" && cf.Username == "" && cf.Password == "" && cf.ClientId == "" {
		return nil, fmt.Errorf("mqtt配置信息不完整")
	}

	clientId := fmt.Sprintf("%s_%s", cf.ClientId, lib.RandString(4))
	connOpts = mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", cf.Address))
	connOpts.SetUsername(cf.Username)
	connOpts.SetPassword(cf.Password)
	connOpts.SetClientID(clientId)

	if cf.ClientCertificateKey != "" {
		connOpts.AddBroker(fmt.Sprintf("ssl://%s", cf.Address))
		tlsConfig := NewTlsConfig(cf.ClientCertificateKey, cf.ClientCertificateCert)
		connOpts.SetTLSConfig(tlsConfig)
	}

	connOpts.SetKeepAlive(cf.KeepAliveDuration * time.Second)
	connOpts.OnConnect = connectHandler            // 连接成功回调
	connOpts.OnConnectionLost = connectLostHandler // 连接断开回调
	connOpts.SetAutoReconnect(true)                // 设置自动重连
	connOpts.SetConnectRetryInterval(time.Second * 5)

	connOpts.OnReconnecting = func(c mqtt.Client, o *mqtt.ClientOptions) {
		log.Debug("尝试重新连接mqtt服务中...")
		c.Connect()
	}

	return
}

// connectHandler 连接成功回调
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Debug("连接MQTT服务成功")
}

// connectLostHandler 连接断开回调
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Debug("MQTT服务连接已断开 %v\n", err)
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
