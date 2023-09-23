# SagooIOT 网关基础服务库代码

开发SagooIOT专属网关时，可以引用此项目，以便快速开发。


## SagooIOT的网关开发说明

```go
    go get -u github.com/sagoo-cloud/iotgateway
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

	//定义事件返回数据
	var eventData = make(map[string]interface{})
	eventData["XXX字段1"] = "XXX值1"
	eventData["XXX字段2"] = "XXX值2"

	var eventDataList = make(map[string]interface{})
	eventDataList["XXX事件标识字串"] = eventData


	//推送数据到mqtt
	out := g.Map{
		"DeviceKey":     deviceKey,
		"EventDataList": eventDataList,
	}

	//触发向MQTT服务推送数据事件
	event.MustFire(consts.PushAttributeDataToMQTT, out) 

```
**属性数据上报**

触发的是 `consts.PushAttributeDataToMQTT` 事件

```go

	var propertieData = make(map[string]interface{})
	propertieData["XXX字段1"] = "XXX值1"
	propertieData["XXX字段2"] = "XXX值2"
	
	//推送数据
	out := g.Map{
		"DeviceKey":     deviceKey,
		"PropertieDataList": propertieData,
	}
	event.MustFire(consts.PushAttributeDataToMQTT, out)

```

由SagooIOT平台端下发后回复：
触发的是 `consts.PushServiceResDataToMQTT` 事件

```go
			var replyData = make(map[string]interface{})
            replyData["XXX字段1"] = "XXX值1"
            replyData["XXX字段2"] = "XXX值1"
			outData := g.Map{
				"DeviceKey": deviceKey,
				"ReplyData": replyData,
			}
			event.MustFire(consts.PushServiceResDataToMQTT, outData)
```