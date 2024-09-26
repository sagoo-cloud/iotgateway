# SagooIOT 网关基础服务库代码

开发SagooIOT专属网关时，可以引用此项目，以便快速开发。创建一个空的工程，按下面的步骤完成自己专属网关的开发。

详细使用可以参考示例工程：[iotgateway-example](https://github.com/sagoo-cloud/iotgateway-example)


## SagooIOT的网关开发说明

```go
    go get -u github.com/sagoo-cloud/iotgateway
```

## 实现入口程序

参考如下：

```go

package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/sagoo-cloud/iotgateway"
	"github.com/sagoo-cloud/iotgateway/version"
)

// 定义编译时的版本信息
var (
	BuildVersion = "0.0"
	BuildTime    = ""
	CommitID     = ""
)

func main() {
	//初始化日志
	glog.SetDefaultLogger(g.Log())
	//显示版本信息
	version.ShowLogo(BuildVersion, BuildTime, CommitID)
	ctx := gctx.GetInitCtx()

	//需要解析的协议，可以根据需要添加,如果不需要实现自定义解析协议，可以不添加，可以为nil
	chargeProtocol := protocol.ChargeProtocol{}

	//创建网关
	gateway, err := iotgateway.NewGateway(ctx, chargeProtocol)
	if err != nil {
		panic(err)
	}
	//初始化事件
	events.Init()                     
	
	// 初始化个性网关需要实现的其它服务

	//启动网关
	gateway.Start()

}


```


## 实现protocol接口

实现protocol接口处理接收到的数据。在Decode方法中，需要将接收到的数据进行解析，然后返回解析后的数据。在Encode方法中，需要将需要发送的数据进行编码，然后返回编码后的数据。

```go
type ChargeProtocol struct {
}

func (c *ChargeProtocol) Encode(args []byte) (res []byte, err error) {
	return args, nil
}

func (c *ChargeProtocol) Decode(conn net.Conn, buffer []byte) (res []byte, err error) {
    return buffer, nil
}   

```


## 向SagooIOT服务推送数据

在需要推送数据的地方准备好事件相关数据，然后触发推送事件。

**事件数据上报：**

触发的是 `consts.PushAttributeDataToMQTT` 事件


```go

	//准备事件返回数据
	var eventData = make(map[string]interface{})
	eventData["XXX字段1"] = "XXX值1"
	eventData["XXX字段2"] = "XXX值2"

	var eventDataList = make(map[string]interface{})
	eventDataList["XXX事件标识字串"] = eventData

	out := g.Map{
		"DeviceKey":     deviceKey,
		"EventDataList": eventDataList,
	}
	//触发属性上报事件
	event.MustFire(consts.PushAttributeDataToMQTT, out) 

```
**属性数据上报**

触发的是 `consts.PushAttributeDataToMQTT` 事件

```go
	//准备上报的数据
	var propertieData = make(map[string]interface{})
	propertieData["XXX字段1"] = "XXX值1"
	propertieData["XXX字段2"] = "XXX值2"
	
	out := g.Map{
		"DeviceKey":     deviceKey,
		"PropertieDataList": propertieData,
	}
	//触发属性上报事件
	event.MustFire(consts.PushAttributeDataToMQTT, out)

```

## 从SagooIOT平台下发调用回复

在SagooIoT系统中向设备端下发有两种情况，1. 服务下发，2. 属性设置下发。

### 服务下发

如果需要完成SagooIoT端向设备进行服务调用，需要在网关程序中完成订阅服务下发事件。
触发的是 `consts.PushServiceResDataToMQTT` 事件。

一 、在获取到设备key的地方订阅服务下发事件。

```go
		//订阅网关设备服务下发事件
    iotgateway.ServerGateway.SubscribeServiceEvent(传入获取的设备key)

```
二、在对设备进行处理后，需要回复SagooIOT平台。

由SagooIOT平台端下发后回复：
触发的是 `consts.PushServiceResDataToMQTT` 事件

```go
      //准备回复数据
			var replyData = make(map[string]interface{})
            replyData["XXX字段1"] = "XXX值1"
            replyData["XXX字段2"] = "XXX值1"
			outData := g.Map{
				"DeviceKey": deviceKey,
				"ReplyData": replyData,
			}
			//出发回复的事件
			event.MustFire(consts.PushServiceResDataToMQTT, outData)
```

### 属性设置下发

如果需要完成SagooIoT端向设备进行服务调用，需要在网关程序中完成订阅服务下发事件。
触发的是 `consts.PropertySetEvent` 事件。

一 、在获取到设备key的地方订阅服务下发事件。

```go
		//订阅网关设备服务下发事件
    iotgateway.ServerGateway.SubscribeSetEvent(传入获取的设备key)

```
二、在对设备进行处理后，需要回复SagooIOT平台。

由SagooIOT平台端下发后回复：
触发的是 `consts.PushSetResDataToMQTT` 事件

```go
      //准备回复数据
			var replyData = make(map[string]interface{})
            replyData["XXX字段1"] = "XXX值1"
            replyData["XXX字段2"] = "XXX值1"
			outData := g.Map{
				"DeviceKey": deviceKey,
				"ReplyData": replyData,
			}
			//出发回复的事件
			event.MustFire(consts.PushSetResDataToMQTT, outData)
```

### SagooIoT平台接收到回复的数据处理

在SagooIoT平台，对服务下发后，会收到回复数据。需要在对应的功能定义设置输入参数。参数标识与数据类型要与回服务回复的数据保持一致。


## 默认服务下发功能

网关中已经有一些默认的服务下发功能。

### 获取网关版本信息

功能标识：`getGatewayVersion`
功能描述：获取网关版本信息
功能输入参数：无
功能输出参数：

| 参数标识  | 参数名称 | 类型   |
| --------- | -------- | ------ |
| Version   | 版本     | string |
| BuildTime | 编译时间 | string |
| CommitID  | 提交ID   | string |