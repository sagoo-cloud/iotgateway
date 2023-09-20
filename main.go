package main

import (
	"github.com/sagoo-cloud/iotgateway/conf"
	"log"
)

func main() {

	var cf conf.GatewayConfig
	cf.Mqtt.Address = "127.0.0.1:1883"
	cf.Mqtt.Username = "xinjy"
	cf.Mqtt.Password = "xinjy123456"
	cf.Mqtt.ClientId = "sagoo-iot-gateway"
	cf.Mqtt.KeepAliveDuration = 60
	cf.Mqtt.Duration = 60
	cf.Server.Addr = ":60972"
	cf.Server.Duration = 60
	cf.Server.ProductKey = "PPPaaaaa"
	cf.Server.DeviceKey = "ddduuuuu"

	log.Println("TCPServer started listening ", cf.Server.Addr)
	NewGateway(cf).Run()
}
