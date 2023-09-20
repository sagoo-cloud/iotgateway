package main

import (
	"context"
	"github.com/sagoo-cloud/iotgateway/conf"
	"github.com/sagoo-cloud/iotgateway/mqttClient"
	"github.com/sagoo-cloud/iotgateway/network"
	"github.com/sagoo-cloud/iotgateway/vars"
)

type sagooIotGateway struct {
	conf conf.GatewayConfig
}

func NewGateway(c conf.GatewayConfig) *sagooIotGateway {
	s := &sagooIotGateway{}
	s.conf = c
	return &sagooIotGateway{}
}

func (s *sagooIotGateway) Run() (err error) {
	mqttClient.GetMQTTClient(s.conf.Mqtt) //初始化mqtt客户端
	vars.Gateway = s.conf.Server          //初始化网关配置
	//go httpTest.HttpServer()                                    //启动http服务
	network.StartServer(context.Background(), s.conf.Server.Addr, s.conf.Protocol) //启动tcp服务
	return
}
