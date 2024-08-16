package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/sagoo-cloud/iotgateway"
	"github.com/sagoo-cloud/iotgateway/version"
	"sagoo-example-gateway/events"
	"sagoo-example-gateway/trapServer"
)

// 定义编译时的版本信息
var (
	BuildVersion = "0.0"
	BuildTime    = ""
	CommitID     = ""
)

// 请参考：https://iotdoc.sagoo.cn/develop/protocol/iotgateway
func main() {
	glog.SetDefaultLogger(g.Log())
	//显示版本信息
	version.ShowLogo(BuildVersion, BuildTime, CommitID)
	ctx := gctx.GetInitCtx()

	//创建网关,如果是TCP设备，则需要传入协议，如果为mqtt设备，protocol为nil，设备类型在配置文件中进行配置：netType的值
	//gateway, err := iotgateway.NewGateway(ctx, nil)

	protocol := &trapServer.ChargeProtocol{}
	gateway, err := iotgateway.NewGateway(ctx, protocol)
	if err != nil {
		panic(err)
	}

	events.Init()                         //初始化事件
	go trapServer.InitializeAndServe(ctx) // 初始化modbus服务

	//启动网关
	gateway.Start()

}
